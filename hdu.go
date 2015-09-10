// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
)

// HDUType is the type of a Header-Data Unit
type HDUType int

const (
	IMAGE_HDU  HDUType = iota // Primary Array or IMAGE HDU
	ASCII_TBL                 // ASCII table HDU
	BINARY_TBL                // Binary table HDU
	ANY_HDU                   // matches any HDU type
)

func (htype HDUType) String() string {
	switch htype {
	case IMAGE_HDU:
		return "IMAGE"
	case ASCII_TBL:
		return "TABLE"
	case BINARY_TBL:
		return "BINTABLE"
	case ANY_HDU:
		return "ANY_HDU"
	default:
		panic(fmt.Errorf("invalid HDU Type value (%v)", int(htype)))
	}
}

// HDU is a "Header-Data Unit" block
type HDU interface {
	Close() error
	Type() HDUType
	Name() string
	Version() int
	Header() *Header
}

// CopyHDU copies the i-th HDU from the src FITS file into the dst one.
func CopyHDU(dst, src *File, i int) error {
	// FIXME(sbinet)
	// use a more efficient implementation. directly copying raw-bytes
	// instead of decoding/re-encoding them.
	hdu := src.HDU(i)
	return dst.Write(hdu)
}
