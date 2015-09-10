// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"reflect"
)

// typecode represents FITS types
type typecode int

const (
	tcInvalid typecode = iota
	tcByte
	tcBool
	tcString
	tcUint16
	tcUint32
	tcUint64
	tcInt8
	tcInt16
	tcInt32
	tcInt64
	tcFloat32
	tcFloat64
	tcComplex64
	tcComplex128

	tcByteVLA       = -tcByte
	tcBoolVLA       = -tcBool
	tcStringVLA     = -tcString
	tcUint16VLA     = -tcUint16
	tcUint32VLA     = -tcUint32
	tcUint64VLA     = -tcUint64
	tcInt8VLA       = -tcInt8
	tcInt16VLA      = -tcInt16
	tcInt32VLA      = -tcInt32
	tcInt64VLA      = -tcInt64
	tcFloat32VLA    = -tcFloat32
	tcFloat64VLA    = -tcFloat64
	tcComplex64VLA  = -tcComplex64
	tcComplex128VLA = -tcComplex128
)

// Type describes a FITS type and its associated Go type
type Type struct {
	tc     typecode     // FITS typecode
	len    int          // number of elements (slice or array)
	dsize  int          // type size in bytes in main data table
	hsize  int          // type size in bytes in heap area
	gotype reflect.Type // associated go type
}
