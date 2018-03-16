// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
	"math/big"
	"reflect"
)

// Header describes a Header-Data Unit of a FITS file
type Header struct {
	htype  HDUType // type of the HDU
	bitpix int     // character information
	axes   []int   // dimensions of image data array
	cards  []Card  // content of the Header
}

// newHeader creates a new Header.
// no test on the content of the cards is performed.
func newHeader(cards []Card, htype HDUType, bitpix int, axes []int) *Header {
	hdr := &Header{
		htype:  htype,
		bitpix: bitpix,
		axes:   make([]int, len(axes)),
		cards:  make([]Card, 0, len(cards)),
	}
	copy(hdr.axes, axes)
	err := hdr.Append(cards...)
	if err != nil {
		panic(err)
	}
	return hdr
}

// NewHeader creates a new Header from a set of Cards, HDU Type, bitpix and axes.
func NewHeader(cards []Card, htype HDUType, bitpix int, axes []int) *Header {
	hdr := newHeader(cards, htype, bitpix, axes)

	// add (some) mandatory cards (BITPIX, NAXES, AXIS1, AXIS2)
	keys := make(map[string]struct{}, len(cards))
	for i := range hdr.cards {
		card := &hdr.cards[i]
		k := card.Name
		keys[k] = struct{}{}
	}

	dcards := make([]Card, 0, 3)
	if _, ok := keys["BITPIX"]; !ok {
		dcards = append(dcards, Card{
			Name:    "BITPIX",
			Value:   hdr.Bitpix(),
			Comment: "number of bits per data pixel",
		})
	}

	if _, ok := keys["NAXIS"]; !ok {
		dcards = append(dcards, Card{
			Name:    "NAXIS",
			Value:   len(hdr.Axes()),
			Comment: "number of data axes",
		})
	}

	if len(hdr.Axes()) >= 1 {
		if _, ok := keys["NAXIS1"]; !ok {
			dcards = append(dcards, Card{
				Name:    "NAXIS1",
				Value:   hdr.Axes()[0],
				Comment: "length of data axis 1",
			})
		}
	}

	if len(hdr.Axes()) >= 2 {
		if _, ok := keys["NAXIS2"]; !ok {
			dcards = append(dcards, Card{
				Name:    "NAXIS2",
				Value:   hdr.Axes()[1],
				Comment: "length of data axis 2",
			})
		}
	}

	err := hdr.prepend(dcards...)
	if err != nil {
		panic(err)
	}
	return hdr
}

// NewDefaultHeader creates a Header with CFITSIO default Cards, of type IMAGE_HDU and
// bitpix=8, no axes.
func NewDefaultHeader() *Header {
	return NewHeader(
		[]Card{
			{
				Name:    "SIMPLE",
				Value:   true,
				Comment: "file does conform to FITS standard",
			},
			{
				Name:    "BITPIX",
				Value:   8,
				Comment: "number of bits per data pixel",
			},
			{
				Name:    "NAXIS",
				Value:   0,
				Comment: "number of data axes",
			},
			{
				Name:    "NAXIS1",
				Value:   0,
				Comment: "length of data axis 1",
			},
			{
				Name:    "NAXIS2",
				Value:   0,
				Comment: "length of data axis 2",
			},
		},
		IMAGE_HDU,
		8,
		[]int{},
	)
}

// Type returns the Type of this Header
func (hdr *Header) Type() HDUType {
	return hdr.htype
}

// Text returns the header's cards content as 80-byte lines
func (hdr *Header) Text() string {
	const kLINE = 80
	buf := make([]byte, 0, kLINE*len(hdr.cards))
	for i := range hdr.cards {
		card := &hdr.cards[i]
		line, err := makeHeaderLine(card)
		if err != nil {
			panic(fmt.Errorf("fitsio: %v", err))
		}
		buf = append(buf, line...)
	}
	return string(buf)
}

// Append appends a set of Cards to this Header
func (hdr *Header) Append(cards ...Card) error {
	var err error
	keys := make(map[string]struct{}, len(hdr.cards))
	for i := range hdr.cards {
		card := &hdr.cards[i]
		k := card.Name
		keys[k] = struct{}{}
	}

	for _, card := range cards {
		_, dup := keys[card.Name]
		if dup {
			switch card.Name {
			case "COMMENT", "HISTORY", "":
				hdr.cards = append(hdr.cards, card)
				continue
			case "END":
				continue
			default:
				return fmt.Errorf("fitsio: duplicate Card [%s] (value=%v)", card.Name, card.Value)
			}
		}
		rv := reflect.ValueOf(card.Value)
		if rv.IsValid() {
			switch rv.Type().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				card.Value = int(rv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				card.Value = int(rv.Uint())
			case reflect.Float32, reflect.Float64:
				card.Value = rv.Float()
			case reflect.Complex64, reflect.Complex128:
				card.Value = rv.Complex()
			case reflect.String:
				card.Value = card.Value.(string)
			case reflect.Bool:
				card.Value = card.Value.(bool)
			case reflect.Struct:
				switch card.Value.(type) {
				case big.Int:
					// ok
				default:
					return fmt.Errorf(
						"fitsio: invalid value type (%T) for card [%s] (kind=%v)",
						card.Value, card.Name, rv.Type().Kind(),
					)
				}
			default:
				return fmt.Errorf(
					"fitsio: invalid value type (%T) for card [%s] (kind=%v)",
					card.Value, card.Name, rv.Type().Kind(),
				)
			}
		}
		hdr.cards = append(hdr.cards, card)
	}
	return err
}

// prepend prepends a (set of) cards to this Header
func (hdr *Header) prepend(cards ...Card) error {
	var err error
	keys := make(map[string]struct{}, len(hdr.cards))
	for i := range hdr.cards {
		card := &hdr.cards[i]
		k := card.Name
		keys[k] = struct{}{}
	}

	hcards := make([]Card, 0, len(cards))
	for _, card := range cards {
		_, dup := keys[card.Name]
		if dup {
			switch card.Name {
			case "COMMENT", "HISTORY", "":
				hdr.cards = append(hdr.cards, card)
				continue
			case "END":
				continue
			default:
				return fmt.Errorf("fitsio: duplicate Card [%s] (value=%v)", card.Name, card.Value)
			}
		}
		rv := reflect.ValueOf(card.Value)
		if rv.IsValid() {
			switch rv.Type().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				card.Value = int(rv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				card.Value = int(rv.Uint())
			case reflect.Float32, reflect.Float64:
				card.Value = rv.Float()
			case reflect.Complex64, reflect.Complex128:
				card.Value = rv.Complex()
			case reflect.String:
				card.Value = card.Value.(string)
			case reflect.Bool:
				card.Value = card.Value.(bool)
			default:
				return fmt.Errorf(
					"fitsio: invalid value type (%T) for card [%s]",
					card.Value, card.Name,
				)
			}
		}
		hcards = append(hcards, card)
	}

	hdr.cards = append(hcards, hdr.cards...)
	return err
}

// Clear resets the Header to the default state.
func (hdr *Header) Clear() {
	hdr.cards = make([]Card, 0)
	hdr.bitpix = 0
	hdr.axes = make([]int, 0)
}

// get returns the Card (and its index) with name n if it exists.
func (hdr *Header) get(n string) (int, *Card) {
	for i := range hdr.cards {
		c := &hdr.cards[i]
		if n == c.Name {
			return i, c
		}
	}
	return -1, nil
}

// Get returns the Card with name n or nil if it doesn't exist.
// If multiple cards with the same name exist, the first one is returned.
func (hdr *Header) Get(n string) *Card {
	_, card := hdr.get(n)
	return card
}

// Card returns the i-th card.
// Card panics if the index is out of range.
func (hdr *Header) Card(i int) *Card {
	return &hdr.cards[i]
}

// Comment returns the whole comment string for this Header.
func (hdr *Header) Comment() string {
	card := hdr.Get("COMMENT")
	if card != nil {
		return card.Value.(string)
	}
	return ""
}

// History returns the whole history string for this Header.
func (hdr *Header) History() string {
	card := hdr.Get("HISTORY")
	if card != nil {
		return card.Value.(string)
	}
	return ""
}

// Bitpix returns the bitpix value.
func (hdr *Header) Bitpix() int {
	return hdr.bitpix
}

// Axes returns the axes for this Header.
func (hdr *Header) Axes() []int {
	return hdr.axes
}

// Index returns the index of the Card with name n, or -1 if it doesn't exist
func (hdr *Header) Index(n string) int {
	idx, _ := hdr.get(n)
	return idx
}

// Keys returns the name of all the Cards of this Header.
func (hdr *Header) Keys() []string {
	keys := make([]string, 0, len(hdr.cards))
	for i := range hdr.cards {
		key := hdr.cards[i].Name
		switch key {
		case "END", "COMMENT", "HISTORY", "":
			continue
		default:
			keys = append(keys, key)
		}
	}
	return keys
}

// Set modifies the value and comment of a Card with name n.
func (hdr *Header) Set(n string, v interface{}, comment string) {
	card := hdr.Get(n)
	if card == nil {
		hdr.Append(Card{
			Name:    n,
			Value:   v,
			Comment: comment,
		})
	} else {
		card.Value = v
		card.Comment = comment
	}
}
