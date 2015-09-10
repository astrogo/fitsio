// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import "reflect"

type primaryHDU struct {
	imageHDU
}

// Name returns the value of the 'EXTNAME' Card (or "PRIMARY" if none)
func (hdu *primaryHDU) Name() string {
	card := hdu.hdr.Get("EXTNAME")
	if card == nil {
		return "PRIMARY"
	}
	return card.Value.(string)
}

// Version returns the value of the 'EXTVER' Card (or 1 if none)
func (hdu *primaryHDU) Version() int {
	card := hdu.hdr.Get("EXTVER")
	if card == nil {
		return 1
	}
	rv := reflect.ValueOf(card.Value)
	return int(rv.Int())
}

// NewPrimaryHDU creates a new PrimaryHDU with Header hdr.
// If hdr is nil, a default Header will be created.
func NewPrimaryHDU(hdr *Header) (Image, error) {
	var err error

	if hdr == nil {
		hdr = NewDefaultHeader()
	}

	// add default cards (SIMPLE, BITPIX, NAXES, AXIS1, AXIS2)
	keys := make(map[string]struct{}, len(hdr.cards))
	for i := range hdr.cards {
		card := &hdr.cards[i]
		k := card.Name
		keys[k] = struct{}{}
	}

	cards := make([]Card, 0, 3)
	if _, ok := keys["SIMPLE"]; !ok {
		cards = append(cards, Card{
			Name:    "SIMPLE",
			Value:   true,
			Comment: "primary HDU",
		})
	}

	if _, ok := keys["BITPIX"]; !ok {
		cards = append(cards, Card{
			Name:    "BITPIX",
			Value:   hdr.Bitpix(),
			Comment: "number of bits per data pixel",
		})
	}

	if _, ok := keys["NAXIS"]; !ok {
		cards = append(cards, Card{
			Name:    "NAXIS",
			Value:   len(hdr.Axes()),
			Comment: "number of data axes",
		})
	}

	if len(hdr.Axes()) >= 1 {
		if _, ok := keys["NAXIS1"]; !ok {
			cards = append(cards, Card{
				Name:    "NAXIS1",
				Value:   hdr.Axes()[0],
				Comment: "length of data axis 1",
			})
		}
	}

	if len(hdr.Axes()) >= 2 {
		if _, ok := keys["NAXIS2"]; !ok {
			cards = append(cards, Card{
				Name:    "NAXIS2",
				Value:   hdr.Axes()[1],
				Comment: "length of data axis 2",
			})
		}
	}

	phdr := *hdr
	phdr.cards = append(cards, hdr.cards...)
	hdu := &primaryHDU{
		imageHDU{
			hdr: phdr,
			raw: make([]byte, 0),
		},
	}

	return hdu, err
}
