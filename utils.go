// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

// var (
// 	g_debug = false
// )
// func printf(format string, args ...interface{}) (int, error) {
// 	if g_debug {
// 		return fmt.Printf(format, args...)
// 	}
// 	return 0, nil
// }

// alignBlock returns a size adjusted to align at a FITS block size
func alignBlock(sz int) int {
	padding := padBlock(sz)
	return sz + padding
}

// padBlock returns the amount of padding to align to a FITS block size
func padBlock(sz int) int {
	padding := (blockSize - (sz % blockSize)) % blockSize
	return padding
}

// processString is utilized by DecodeHDU to process string-type values in the header
// it uses a 3-state machine to process double single quotes
func processString(s string) (string, int, error) {
	var buf bytes.Buffer

	state := 0
	for i, char := range s {
		quote := (char == '\'')
		switch state {
		case 0:
			if !quote {
				return "", i, fmt.Errorf("fitsio: string does not start with a quote (%q)", s)
			}
			state = 1
		case 1:
			if quote {
				state = 2
			} else {
				buf.WriteRune(char)
				state = 1
			}
		case 2:
			if quote {
				buf.WriteRune(char)
				state = 1
			} else {
				return strings.TrimRight(buf.String(), " "), i, nil
			}
		}
	}
	if s[len(s)-1] == '\'' {
		return strings.TrimRight(buf.String(), " "), len(s), nil
	}
	return "", 0, fmt.Errorf("fitsio: string ends prematurely (%q)", s)
}

// parseHeaderLine parses a 80-byte line from an input header FITS block.
// transliteration of CFITSIO's ffpsvc.
func parseHeaderLine(bline []byte) (*Card, error) {
	var err error
	var card Card

	valpos := 0
	keybeg := 0
	keyend := 0

	const (
		kLINE = 80
	)

	var (
		kHIERARCH = []byte("HIERARCH ")
		kCOMMENT  = []byte("COMMENT ")
		kCONTINUE = []byte("CONTINUE")
		kHISTORY  = []byte("HISTORY ")
		kEND      = []byte("END     ")
		kEMPTY    = []byte("        ")
	)

	if len(bline) != kLINE {
		return nil, fmt.Errorf("fitsio: invalid header line length")
	}

	// support for ESO HIERARCH keywords: find the '='
	if bytes.HasPrefix(bline, kHIERARCH) {
		idx := bytes.Index(bline, []byte("="))
		if idx < 0 {
			// no value indicator
			card.Comment = strings.TrimRight(string(bline[8:]), " ")
			return &card, nil
		}
		valpos = idx + 1 // point after '='
		keybeg = len(kHIERARCH)
		keyend = idx

	} else if len(bline) < 9 ||
		bytes.HasPrefix(bline, kCOMMENT) ||
		bytes.HasPrefix(bline, kCONTINUE) ||
		bytes.HasPrefix(bline, kHISTORY) ||
		bytes.HasPrefix(bline, kEND) ||
		bytes.HasPrefix(bline, kEMPTY) ||
		!bytes.HasPrefix(bline[8:], []byte("= ")) { // no '= ' in cols 9-10

		// no value, so the comment extends from cols 9 - 80
		card.Comment = strings.TrimRight(string(bline[8:]), " ")

		if bytes.HasPrefix(bline, kCOMMENT) {
			card.Name = "COMMENT"
		} else if bytes.HasPrefix(bline, kCONTINUE) {
			card.Name = "CONTINUE"
			str := strings.TrimSpace(string(bline[len(kCONTINUE):]))
			value, _, err := processString(str)
			if err != nil {
				return nil, err
			}
			card.Comment = value
			return &card, nil

		} else if bytes.HasPrefix(bline, kHISTORY) {
			card.Name = "HISTORY"
		} else if bytes.HasPrefix(bline, kEND) {
			card.Name = "END"
		} else if bytes.HasPrefix(bline, kEMPTY) ||
			!bytes.HasPrefix(bline[8:], []byte("= ")) {
			card.Name = ""
		}

		return &card, nil
	} else {
		valpos = 10
		keybeg = 0
		keyend = 8
	}

	card.Name = strings.TrimSpace(string(bline[keybeg:keyend]))

	// find number of leading blanks
	nblanks := 0
	for _, c := range bline[valpos:] {
		if c != ' ' {
			break
		}
		nblanks += 1
	}

	if nblanks+valpos == len(bline) {
		// the absence of a value string is legal and simply indicates
		// that the keyword value is undefined.
		// don't write an error message in this case
		return &card, nil
	}

	i := valpos + nblanks
	switch bline[i] {
	case '/': // start of the comment
		i += 1
	case '\'': // quoted string value ?
		str, idx, err := processString(string(bline[i:]))
		if err != nil {
			return nil, err
		}
		switch {
		case len(str) <= 69: // don't exceed 70-char null-terminated string length
			card.Value = str
		case len(str) > 69:
			card.Value = str[:70]
		}
		i += idx

	case '(': // a complex value
		idx := bytes.IndexByte(bline[i:], ')')
		if idx < 0 {
			return nil, fmt.Errorf("fitsio: complex keyword missing closing ')' (%q)", string(bline))
		}
		var x, y float64
		str := strings.TrimSpace(string(bline[i : i+idx+1]))
		_, err = fmt.Sscanf(str, "(%f,%f)", &x, &y)
		if err != nil {
			return nil, err
		}
		card.Value = complex(x, y)
		i += idx + 1

	default: // integer, float or logical FITS value string
		v0 := bline[i]
		value := ""

		// find the end of the token
		if valend := bytes.Index(bline[i:], []byte(" /")); valend < 0 {
			value = string(bline[i:])
		} else {
			value = string(bline[i : i+valend])
		}
		i += len(value)

		if (v0 >= '0' && v0 <= '9') || v0 == '+' || v0 == '-' {
			value = strings.TrimSpace(value)
			if strings.ContainsAny(value, ".DE") {
				value = strings.Replace(value, "D", "E", 1) // converts D type floats to E type
				x, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, err
				}
				card.Value = x
			} else {
				x, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					switch err := err.(type) {
					case *strconv.NumError:
						// try math/big.Int
						if err.Err == strconv.ErrRange {
							var x big.Int
							_, err := fmt.Sscanf(value, "%v", &x)
							if err != nil {
								return nil, err
							}
							card.Value = x
						}
					default:
						return nil, err
					}
				} else {
					card.Value = int(x)
				}
			}
		} else if v0 == 'T' {
			card.Value = true
		} else if v0 == 'F' {
			card.Value = false
		} else {
			return nil, fmt.Errorf("fitsio: invalid card line (%q)", string(bline))
		}
	}

	idx := bytes.IndexByte(bline[i:], '/')
	if idx < 0 {
		// no comment
		return &card, err
	}

	com := bline[i+idx+1:]
	card.Comment = strings.TrimSpace(string(com))
	return &card, err
}

// makeHeaderLine makes a 80-byte line (or more) for a header FITS block from a Card.
// transliterated from CFITSIO's ffmkky.
func makeHeaderLine(card *Card) ([]byte, error) {
	var err error
	const kLINE = 80
	var (
		kCONTINUE = []byte("CONTINUE")
	)

	buf := new(bytes.Buffer)
	buf.Grow(kLINE)

	if card == nil {
		return nil, fmt.Errorf("fitsio: nil Card")
	}

	switch card.Name {
	case "", "COMMENT", "HISTORY":
		str := card.Comment
		vlen := len(str)
		for i := 0; i < vlen; i += 72 {
			end := i + 72
			if end > vlen {
				end = vlen
			}
			_, err = fmt.Fprintf(buf, "%-8s%-72s", card.Name, str[i:end])
			if err != nil {
				return nil, err
			}
		}
		return buf.Bytes(), err
	case "END":
		_, err = fmt.Fprintf(buf, "%-80s", "END")
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), err
	}

	klen := len(card.Name)

	if klen <= 8 && verifyCardName(card) == nil {
		// a normal FITS keyword
		_, err = fmt.Fprintf(buf, "%-8s= ", card.Name)
		if err != nil {
			return nil, err
		}
		klen = 10
	} else {
		// use the ESO HIERARCH convention for longer keyword names

		if strings.Contains(card.Name, "=") {
			return nil, fmt.Errorf(
				"fitsio: illegal keyword name. contains an equal sign [%s]",
				card.Name,
			)
		}
		key := card.Name
		// dont repeat HIERARCH if the keyword already contains it
		if !strings.HasPrefix(card.Name, "HIERARCH ") &&
			!strings.HasPrefix(card.Name, "hierarch ") {
			key = "HIERARCH " + card.Name
		}
		n, err := fmt.Fprintf(buf, "%s= ", key)
		if err != nil {
			return nil, err
		}
		klen = n
	}

	if card.Value == nil {
		// this case applies to normal keywords only
		if klen == 10 {
			// keywords with no value have no '='
			buf.Bytes()[8] = ' '
			if card.Comment != "" {
				comment := " / " + card.Comment
				max := len(comment)
				if max > kLINE-klen {
					max = kLINE - klen
				}
				_, err = fmt.Fprintf(buf, "%s", comment[:max])
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		buflen := buf.Len()
		//valstr := ""
		n := 0
		switch v := card.Value.(type) {
		case string:
			vstr := "''"
			if v != "" {
				vstr = fmt.Sprintf("'%-8s'", v)
			}
			if len(vstr) < kLINE-buflen {
				n, err = fmt.Fprintf(buf, "%-20s", vstr)
				if err != nil {
					return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
				}
			} else {
				// string too long.
				// use CONTINUE blocks.
				// replace last character of string with '&'
				ampersand := len("&")
				quotes := len("''")
				spacesz := len("  ")
				sz := kLINE - buflen - ampersand - quotes
				vstr = fmt.Sprintf("'%-8s'", v[:sz]+"&")
				n, err = fmt.Fprintf(buf, "%-20s", vstr)
				if err != nil {
					return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
				}
				contlen := len(kCONTINUE)
				blocksz := kLINE - contlen - ampersand - quotes - spacesz
				for i := sz; i < len(v); i += blocksz {
					end := i + blocksz
					amper := "&"
					if end > len(v) {
						end = len(v)
						amper = ""
					}
					vv := v[i:end]
					vstr := fmt.Sprintf("'%-8s'", vv+amper)
					n, err = fmt.Fprintf(buf, "%s  %-20s", string(kCONTINUE), vstr)
					if err != nil {
						return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
					}
				}
				// fill buffer up to 80-byte mark so any remaining comment
				// will have to be handled by a separate 'COMMENT' line
				n = buf.Len()
				align80 := (kLINE - (n % kLINE)) % kLINE
				if align80 > 0 {
					_, err = buf.Write(bytes.Repeat([]byte(" "), align80))
					if err != nil {
						return nil, err
					}
				}

				n = 0
				buflen = buf.Len() % kLINE
			}

		case bool:
			vv := "F"
			if v {
				vv = "T"
			}
			n, err = fmt.Fprintf(buf, "%20s", vv)
			if err != nil {
				return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
			}

		case int:
			n, err = fmt.Fprintf(buf, "%20d", v)
			if err != nil {
				return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
			}

		case float64:
			n, err = fmt.Fprintf(buf, "%20f", v)
			if err != nil {
				return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
			}

		case complex128:
			n, err = fmt.Fprintf(buf, "(%10f,%10f)", real(v), imag(v))
			if err != nil {
				return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
			}

		case big.Int:
			n, err = fmt.Fprintf(buf, "%s", v.String())
			if err != nil {
				return nil, fmt.Errorf("fitsio: error writing card value [%s]: %v", card.Name, err)
			}

		default:
			panic(fmt.Errorf("fitsio: invalid card value [%s]: %#v (%T)", card.Name, v, v))
		}

		if n+buflen > kLINE {
			return nil, fmt.Errorf("fitsio: value-string too big (%d) for card [%s]: %v\nbuf=|%s|",
				n, card.Name, card.Value, string(buf.Bytes()),
			)
		}

		buflen = buf.Len() % kLINE

		comment := " / " + card.Comment
		max := len(comment)
		// if card.Comment == "my-comment" {
		// 	fmt.Printf("max=%d\n", max)
		// 	fmt.Printf("buf=%d\n", buflen)
		// 	fmt.Printf("buf=%d\n", buf.Len())
		// 	fmt.Printf("dif=%d\n", kLINE-buflen)
		// }

		if max > kLINE-buflen || (buf.Len() > kLINE && (buf.Len()%kLINE) == 0) {
			// append a 'COMMENT' line
			if buflen > 0 {
				_, err = buf.Write(bytes.Repeat([]byte(" "), kLINE-buflen))
				if err != nil {
					return nil, err
				}
			}
			comline, err := makeHeaderLine(&Card{Name: "COMMENT", Comment: card.Comment})
			if err != nil {
				return nil, err
			}
			_, err = buf.Write(comline)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = fmt.Fprintf(buf, "%s", comment[:max])
			if err != nil {
				return nil, err
			}
		}

	}

	n := buf.Len()
	align80 := (kLINE - (n % kLINE)) % kLINE
	if align80 > 0 {
		_, err = buf.Write(bytes.Repeat([]byte(" "), align80))
	}
	return buf.Bytes(), err
}

// verifyCardName verifies a Card name conforms to the FITS standard.
// Must contain only capital letters, digits, minus or underscore chars.
// Trailing spaces are allowed.
func verifyCardName(card *Card) error {
	var err error
	spaces := false

	max := len(card.Name)
	if max > 8 {
		max = 8
	}

	for idx, c := range card.Name {
		switch {
		case (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_':
			if spaces {
				return fmt.Errorf("fitsio: card name contains embedded space(s): %q", card.Name)
			}
		case c == ' ':
			spaces = true
		default:
			return fmt.Errorf(
				"fitsio: card name contains illegal character %q (idx=%d)",
				card.Name, idx,
			)
		}
	}

	return err
}

// typeFromForm returns a FITS Type corresponding to a FITS TFORM string
func typeFromForm(form string, htype HDUType) (Type, error) {
	var err error
	var typ Type

	switch htype {
	case BINARY_TBL:
		j := strings.IndexAny(form, "PQABCDEIJKLMX")
		if j < 0 {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s)", form)
		}
		repeat := 1
		if j > 0 {
			r, err := strconv.ParseInt(form[:j], 10, 32)
			if err != nil {
				return typ, fmt.Errorf("fitsio: invalid TFORM format (%s)", form)
			}
			repeat = int(r)
		}
		slice := false
		dsize := 0
		hsize := 0
		switch form[j] {
		case 'P':
			j += 1
			slice = true
			dsize = 2 * 4
		case 'Q':
			j += 1
			slice = true
			dsize = 2 * 8
		}
		tc, ok := g_fits2tc[BINARY_TBL][form[j]]
		if !ok {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s) (no typecode found)", form)
		}
		rt, ok := g_fits2go[BINARY_TBL][form[j]]
		if !ok {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s) (no Type found)", form)
		}

		elemsz := 0
		switch form[j] {
		case 'A':
			elemsz = repeat
			repeat = 1
		case 'X':
			elemsz = 1
			const nbits = 8
			sz := repeat + (nbits-(repeat%nbits))%nbits
			repeat = sz / nbits

		case 'L', 'B':
			elemsz = 1
		case 'I':
			elemsz = 2
		case 'J', 'E':
			elemsz = 4
		case 'K', 'D', 'C':
			elemsz = 8
		case 'M':
			elemsz = 16
		}

		switch slice {
		case true:
			hsize = elemsz
			typ = Type{
				tc:     -tc,
				len:    repeat,
				dsize:  dsize,
				hsize:  hsize,
				gotype: reflect.SliceOf(rt),
			}

		case false:
			dsize = elemsz
			if repeat > 1 {
				typ = Type{
					tc:     tc,
					len:    repeat,
					dsize:  dsize,
					hsize:  hsize,
					gotype: reflect.ArrayOf(repeat, rt),
				}
			} else {
				typ = Type{
					tc:     tc,
					len:    repeat,
					dsize:  dsize,
					hsize:  hsize,
					gotype: rt,
				}
			}
		}

		if typ.dsize*typ.len == 0 {
			if form != "0A" {
				return typ, fmt.Errorf("fitsio: invalid dtype! form=%q typ=%#v\n", form, typ)
			}
		}

	case ASCII_TBL:
		// fmt.Printf("### form %q\n", form)
		j := strings.IndexAny(form, "ADEFI")
		if j < 0 {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s)", form)
		}
		j = strings.Index(form, ".")
		if j == -1 {
			j = len(form)
		}
		repeat := 1
		r, err := strconv.ParseInt(form[1:j], 10, 32)
		if err != nil {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s)", form)
		}
		repeat = int(r)

		tc, ok := g_fits2tc[ASCII_TBL][form[0]]
		if !ok {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s) (no typecode found)", form)
		}
		rt, ok := g_fits2go[ASCII_TBL][form[0]]
		if !ok {
			return typ, fmt.Errorf("fitsio: invalid TFORM format (%s) (no Type found)", form)
		}

		dsize := 0
		hsize := 0

		switch form[0] {
		case 'A':
			dsize = repeat
			repeat = 1
		case 'I':
			dsize = repeat
			repeat = 1
		case 'D', 'E':
			dsize = repeat
			repeat = 1
		case 'F':
			dsize = repeat
			repeat = 1
		}

		typ = Type{
			tc:     tc,
			len:    repeat,
			dsize:  dsize,
			hsize:  hsize,
			gotype: rt,
		}
		// fmt.Printf(">>> %#v (%v)\n", typ, typ.gotype.Name())

		if typ.dsize*typ.len == 0 {
			if form != "0A" {
				return typ, fmt.Errorf("fitsio: invalid dtype! form=%q typ=%#v\n", form, typ)
			}
		}
	}

	return typ, err
}

// txtfmtFromForm returns a suitable go-fmt format from a FITS TFORM
func txtfmtFromForm(form string) string {
	var format string
	var code rune
	m := -1
	w := 14

	fmt.Sscanf(form, "%c%d.%d", &code, &w, &m)

	switch form[0] {
	case 'A':
		format = fmt.Sprintf("%%%d.%ds", w, w) // Aw -> %ws
	case 'I':
		format = fmt.Sprintf("%%%dd", w) // Iw -> %wd
	case 'B':
		format = fmt.Sprintf("%%%db", w) // Bw -> %wb, binary
	case 'O':
		format = fmt.Sprintf("%%%do", w) // Ow -> %wo, octal
	case 'Z':
		format = fmt.Sprintf("%%%dX", w) // Zw -> %wX, hexadecimal
	case 'F':
		if m != -1 {
			format = fmt.Sprintf("%%%d.%df", w, m) // Fw.d -> %w.df
		} else {
			format = fmt.Sprintf("%%%df", w) // Fw -> %wf
		}
	case 'E', 'D':
		if m != -1 {
			format = fmt.Sprintf("%%%d.%de", w, m) // Fw.d -> %w.df
		} else {
			format = fmt.Sprintf("%%%de", w) // Ew -> %we
		}
	case 'G':
		if m != -1 {
			format = fmt.Sprintf("%%%d.%dg", w, m) // Fw.d -> %w.df
		} else {
			format = fmt.Sprintf("%%%dg", w) // Gw -> %wg
		}
	}
	return format
}

// formFromGoType returns a suitable FITS TFORM string from a reflect.Type
func formFromGoType(rt reflect.Type, htype HDUType) string {
	hdr := ""
	var t reflect.Type
	switch rt.Kind() {
	case reflect.Array:
		hdr = fmt.Sprintf("%d", rt.Len())
		t = rt.Elem()
	case reflect.Slice:
		hdr = "Q"
		t = rt.Elem()
	default:
		t = rt
	}

	dict, ok := g_gotype2FITS[t.Kind()]
	if !ok {
		return ""
	}

	form, ok := dict[htype]
	if !ok {
		return ""
	}

	return hdr + form
}

var g_gotype2FITS = map[reflect.Kind]map[HDUType]string{

	reflect.Bool: {
		ASCII_TBL:  "",
		BINARY_TBL: "L",
	},

	reflect.Int: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "K",
	},

	reflect.Int8: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "B",
	},

	reflect.Int16: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "I",
	},

	reflect.Int32: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "J",
	},

	reflect.Int64: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "K",
	},

	reflect.Uint: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "V",
	},

	reflect.Uint8: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "B",
	},

	reflect.Uint16: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "U",
	},

	reflect.Uint32: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "V",
	},

	reflect.Uint64: {
		ASCII_TBL:  "I4",
		BINARY_TBL: "V",
	},

	reflect.Uintptr: {
		ASCII_TBL:  "",
		BINARY_TBL: "",
	},

	reflect.Float32: {
		ASCII_TBL:  "E26.17", // must write as float64 since we can only read as such
		BINARY_TBL: "E",
	},

	reflect.Float64: {
		ASCII_TBL:  "E26.17",
		BINARY_TBL: "D",
	},

	reflect.Complex64: {
		ASCII_TBL:  "",
		BINARY_TBL: "C",
	},

	reflect.Complex128: {
		ASCII_TBL:  "",
		BINARY_TBL: "M",
	},

	reflect.Array: {
		ASCII_TBL:  "",
		BINARY_TBL: "",
	},

	reflect.Slice: {
		ASCII_TBL:  "",
		BINARY_TBL: "",
	},

	reflect.String: {
		ASCII_TBL:  "A80",
		BINARY_TBL: "80A",
	},
}

var g_fits2go = map[HDUType]map[byte]reflect.Type{
	ASCII_TBL: {
		'A': reflect.TypeOf((*string)(nil)).Elem(),
		'I': reflect.TypeOf((*int)(nil)).Elem(),
		'E': reflect.TypeOf((*float64)(nil)).Elem(),
		'D': reflect.TypeOf((*float64)(nil)).Elem(),
		'F': reflect.TypeOf((*float64)(nil)).Elem(),
	},

	BINARY_TBL: {
		'A': reflect.TypeOf((*string)(nil)).Elem(),
		'B': reflect.TypeOf((*byte)(nil)).Elem(),
		'L': reflect.TypeOf((*bool)(nil)).Elem(),
		'I': reflect.TypeOf((*int16)(nil)).Elem(),
		'J': reflect.TypeOf((*int32)(nil)).Elem(),
		'K': reflect.TypeOf((*int64)(nil)).Elem(),
		'E': reflect.TypeOf((*float32)(nil)).Elem(),
		'D': reflect.TypeOf((*float64)(nil)).Elem(),
		'C': reflect.TypeOf((*complex64)(nil)).Elem(),
		'M': reflect.TypeOf((*complex128)(nil)).Elem(),
		'X': reflect.TypeOf((*byte)(nil)).Elem(),
	},
}

var g_fits2tc = map[HDUType]map[byte]typecode{
	ASCII_TBL: {
		'A': tcString,
		'I': tcInt64,
		'E': tcFloat64,
		'D': tcFloat64,
		'F': tcFloat64,
	},

	BINARY_TBL: {
		'A': tcString,
		'B': tcByte,
		'L': tcBool,
		'I': tcInt16,
		'J': tcInt32,
		'K': tcInt64,
		'E': tcFloat32,
		'D': tcFloat64,
		'C': tcComplex64,
		'M': tcComplex128,
		'X': tcByte,
	},
}
