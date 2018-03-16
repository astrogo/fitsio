// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

type Decoder interface {
	DecodeHDU() (HDU, error)
}

// NewDecoder creates a new Decoder according to the capabilities of the underlying io.Reader
func NewDecoder(r io.Reader) Decoder {
	// FIXME(sbinet)
	// if rr, ok := r.(io.ReadSeeker); ok {
	// 	return &seekDecoder{r: rr}
	// }
	return &streamDecoder{r: r}
}

// streamDecoder is a decoder which can not perform random access
// into the underlying Reader
type streamDecoder struct {
	r io.Reader
}

func (dec *streamDecoder) DecodeHDU() (HDU, error) {
	var err error
	var hdu HDU

	cards := make(map[string]int, 30)
	slice := make([]Card, 0, 1)

	get_card := func(k string) (Card, bool) {
		i, ok := cards[k]
		if ok {
			return slice[i], ok
		}
		return Card{}, ok
	}

	add_card := func(c *Card) {
		n := c.Name
		if n == "COMMENT" || n == "HISTORY" || n == "" {
			slice = append(slice, *c)
			return
		}
		// For compatibility with C FITSIO, silently swallow duplicate card keys.
		// See:
		//  https://github.com/astrogo/fitsio/issues/38
		//
		// _, dup := cards[n]
		// if dup {
		// 	panic(fmt.Errorf("fitsio: duplicate keyword [%s]", n))
		// }
		cards[n] = len(slice)
		slice = append(slice, *c)
	}

	axes := []int{}
	buf := make([]byte, blockSize)

	iblock := -1
blocks_loop:
	for {
		iblock += 1
		_, err = io.ReadFull(dec.r, buf)
		if err != nil {
			return nil, err
		}

		// each FITS header block is comprised of up to 36 80-byte lines
		const maxlines = 36
		for i := 0; i < maxlines; i++ {
			card, err := parseHeaderLine(buf[i*80 : (i+1)*80])
			if err != nil {
				return nil, err
			}
			if card.Name == "CONTINUE" {
				idx := len(slice) - 1
				last := slice[idx]
				str := last.Value.(string)
				if len(str) > 0 {
					last.Value = str[:len(str)-1] + card.Comment
				}
				slice[idx] = last
				continue
			}
			add_card(card)
			if card.Name == "END" {
				break
			}
		}

		_, ends := get_card("END")
		if ends {
			card, ok := get_card("NAXIS")
			if ok {
				n := card.Value.(int)
				axes = make([]int, n)
				for i := 0; i < n; i++ {
					k := fmt.Sprintf("NAXIS%d", i+1)
					c, ok := get_card(k)
					if !ok {
						return nil, fmt.Errorf("fitsio: missing '%s' key", k)
					}
					axes[i] = c.Value.(int)
				}
			}
			break blocks_loop
		}
	}

	htype, primary, err := hduTypeFrom(slice)
	if err != nil {
		return nil, err
	}

	bitpix := 0
	if card, ok := get_card("BITPIX"); ok {
		bitpix = int(reflect.ValueOf(card.Value).Int())
	} else {
		return nil, fmt.Errorf("fitsio: missing 'BITPIX' card")
	}

	hdr := NewHeader(slice, htype, bitpix, axes)
	switch htype {
	case IMAGE_HDU:
		var data []byte
		data, err = dec.loadImage(hdr)
		if err != nil {
			return nil, fmt.Errorf("fitsio: error loading image: %v", err)
		}

		switch primary {
		case true:
			hdu = &primaryHDU{
				imageHDU: imageHDU{
					hdr: *hdr,
					raw: data,
				},
			}
		case false:
			hdu = &imageHDU{
				hdr: *hdr,
				raw: data,
			}
		}

	case BINARY_TBL:
		hdu, err = dec.loadTable(hdr, htype)
		if err != nil {
			return nil, fmt.Errorf("fitsio: error loading binary table: %v", err)
		}

	case ASCII_TBL:
		hdu, err = dec.loadTable(hdr, htype)
		if err != nil {
			return nil, fmt.Errorf("fitsio: error loading ascii table: %v", err)
		}

	case ANY_HDU:
		fallthrough
	default:
		return nil, fmt.Errorf("fitsio: invalid HDU Type (%v)", htype)
	}

	return hdu, err
}

func (dec *streamDecoder) loadImage(hdr *Header) ([]byte, error) {
	var err error
	var buf []byte

	nelmts := 1
	for _, dim := range hdr.Axes() {
		nelmts *= dim
	}

	if len(hdr.Axes()) <= 0 {
		nelmts = 0
	}

	pixsz := hdr.Bitpix() / 8
	if pixsz < 0 {
		pixsz = -pixsz
	}

	buf = make([]byte, nelmts*pixsz)
	if nelmts == 0 {
		return buf, nil
	}

	n, err := io.ReadFull(dec.r, buf)
	if err != nil {
		return nil, fmt.Errorf("fitsio: error reading %d bytes (got %d): %v", len(buf), n, err)
	}

	// data array is also aligned at 2880-bytes blocks
	pad := padBlock(n)
	if pad > 0 {
		if n, err := io.CopyN(ioutil.Discard, dec.r, int64(pad)); err != nil {
			return nil, fmt.Errorf("fitsio: error reading %d bytes (got %d): %v", pad, n, err)
		}
	}

	return buf, err
}

func (dec *streamDecoder) loadTable(hdr *Header, htype HDUType) (*Table, error) {
	var err error
	var table *Table

	isbinary := true
	switch htype {
	case ASCII_TBL:
		isbinary = false
	case BINARY_TBL:
		isbinary = true
	default:
		return nil, fmt.Errorf("fitsio: invalid HDU type (%v)", htype)
	}

	rowsz := hdr.Axes()[0]
	nrows := int64(hdr.Axes()[1])
	ncols := 0
	if card := hdr.Get("TFIELDS"); card != nil && card.Value != nil {
		ncols = card.Value.(int)
	}

	datasz := int(nrows) * rowsz
	heapsz := 0
	if card := hdr.Get("PCOUNT"); card != nil && card.Value != nil {
		heapsz = card.Value.(int)
	}

	blocksz := alignBlock(datasz + heapsz)

	block := make([]byte, blocksz)
	n, err := io.ReadFull(dec.r, block)
	if err != nil {
		return nil, fmt.Errorf("fitsio: error reading %d bytes (got %d): %v", len(block), n, err)
	}

	gapsz := 0
	if card := hdr.Get("THEAP"); card != nil && card.Value != nil {
		gapsz = card.Value.(int)
	}

	data := block[:datasz]
	heap := block[datasz+gapsz:]

	cols := make([]Column, ncols)
	colidx := make(map[string]int, ncols)

	get := func(str string, ii int) *Card {
		return hdr.Get(fmt.Sprintf(str+"%d", ii+1))
	}

	offset := 0
	for i := 0; i < ncols; i++ {
		col := &cols[i]

		switch htype {
		case BINARY_TBL:
			col.write = col.writeBin
			col.read = col.readBin
		case ASCII_TBL:
			col.write = col.writeTxt
			col.read = col.readTxt
		default:
			return nil, fmt.Errorf("fitsio: invalid HDUType (%v)", htype)
		}

		col.offset = offset

		card := get("TTYPE", i)
		col.Name = card.Value.(string)

		card = get("TFORM", i)
		if card == nil {
			return nil, fmt.Errorf("fitsio: missing 'TFORM%d' for column '%s'", i+1, col.Name)
		} else {
			col.Format = card.Value.(string)
		}

		card = get("TUNIT", i)
		if card != nil && card.Value != nil {
			col.Unit = card.Value.(string)
		}

		card = get("TNULL", i)
		if card != nil && card.Value != nil {
			col.Null = fmt.Sprintf("%v", card.Value)
		}

		card = get("TSCAL", i)
		if card != nil && card.Value != nil {
			switch vv := card.Value.(type) {
			case float64:
				col.Bscale = vv
			case int64:
				col.Bscale = float64(vv)
			case int:
				col.Bscale = float64(vv)
			default:
				return nil, fmt.Errorf("fitsio: unhandled type [%T]", vv)
			}
		} else {
			col.Bscale = 1.0
		}

		card = get("TZERO", i)
		if card != nil && card.Value != nil {
			switch vv := card.Value.(type) {
			case float64:
				col.Bzero = vv
			case int64:
				col.Bzero = float64(vv)
			case int:
				col.Bzero = float64(vv)
			default:
				return nil, fmt.Errorf("fitsio: unhandled type [%T]", vv)
			}
		} else {
			col.Bzero = 0.0
		}

		card = get("TDISP", i)
		if card != nil && card.Value != nil {
			col.Display = card.Value.(string)
		}

		card = get("TDIM", i)
		if card != nil && card.Value != nil {
			dims := card.Value.(string)
			dims = strings.Replace(dims, "(", "", -1)
			dims = strings.Replace(dims, ")", "", -1)
			toks := make([]string, 0)
			for _, tok := range strings.Split(dims, ",") {
				tok = strings.Trim(tok, " \t\n")
				if tok == "" {
					continue
				}
				toks = append(toks, tok)
			}
			col.Dim = make([]int64, 0, len(toks))
			for _, tok := range toks {
				dim, err := strconv.ParseInt(tok, 10, 64)
				if err != nil {
					return nil, err
				}
				col.Dim = append(col.Dim, dim)
			}
		}

		card = get("TBCOL", i)
		if card != nil && card.Value != nil {
			col.Start = int64(card.Value.(int))
		}

		col.dtype, err = typeFromForm(col.Format, htype)
		if err != nil {
			return nil, err
		}
		offset += col.dtype.dsize * col.dtype.len
		if htype == ASCII_TBL {
			col.txtfmt = txtfmtFromForm(col.Format)
		}
		colidx[col.Name] = i
	}

	table = &Table{
		hdr:    *hdr,
		binary: isbinary,
		data:   data,
		heap:   heap,
		rowsz:  rowsz,
		nrows:  nrows,
		cols:   cols,
		colidx: colidx,
	}

	return table, err
}

// seekDecoder is a decoder which can perform random access into
// the underlying Reader
type seekDecoder struct {
	r io.ReadSeeker
}

func (dec *seekDecoder) DecodeHDU() (HDU, error) {
	panic("not implemented")
}

func hduTypeFrom(cards []Card) (HDUType, bool, error) {
	var err error
	var htype HDUType = -1
	var primary bool

	keys := make([]string, 0, len(cards))
	for _, card := range cards {
		keys = append(keys, card.Name)
		switch card.Name {
		case "SIMPLE":
			primary = true
			return IMAGE_HDU, primary, nil
		case "XTENSION":
			str := card.Value.(string)
			switch str {
			case "IMAGE":
				htype = IMAGE_HDU
			case "TABLE":
				htype = ASCII_TBL
			case "BINTABLE":
				htype = BINARY_TBL
			case "ANY", "ANY_HDU":
				htype = ANY_HDU
			default:
				return htype, primary, fmt.Errorf("fitsio: invalid 'XTENSION' value: %q", str)
			}

			return htype, primary, err
		}
	}

	return htype, primary, fmt.Errorf("fitsio: invalid header (missing 'SIMPLE' or 'XTENSION' card): keys=%v", keys)
}
