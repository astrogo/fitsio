// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/gonuts/binary"
)

// Image represents a FITS image
type Image interface {
	HDU
	Read(ptr interface{}) error
	Write(ptr interface{}) error
	Raw() []byte

	freeze() error
}

// imageHDU is a Header-Data Unit extension holding an image as data payload
type imageHDU struct {
	hdr Header
	raw []byte
}

// NewImage creates a new Image with bitpix size for the pixels and axes as its axes
func NewImage(bitpix int, axes []int) *imageHDU {
	hdr := NewHeader(nil, IMAGE_HDU, bitpix, axes)
	return &imageHDU{
		hdr: *hdr,
		raw: make([]byte, 0),
	}
}

// Close closes this HDU, cleaning up cycles (if any) for garbage collection
func (img *imageHDU) Close() error {
	return nil
}

// Header returns the Header part of this HDU block.
func (img *imageHDU) Header() *Header {
	return &img.hdr
}

// Type returns the Type of this HDU
func (img *imageHDU) Type() HDUType {
	return img.hdr.Type()
}

// Name returns the value of the 'EXTNAME' Card.
func (img *imageHDU) Name() string {
	card := img.hdr.Get("EXTNAME")
	if card == nil {
		return ""
	}
	return card.Value.(string)
}

// Version returns the value of the 'EXTVER' Card (or 1 if none)
func (img *imageHDU) Version() int {
	card := img.hdr.Get("EXTVER")
	if card == nil {
		return 1
	}
	return card.Value.(int)
}

// Raw returns the raw bytes which make the image
func (img *imageHDU) Raw() []byte {
	return img.raw
}

// Read reads the image data into ptr
func (img *imageHDU) Read(ptr interface{}) error {
	var err error
	if img.raw == nil {
		// FIXME(sbinet): load data from file
		panic(fmt.Errorf("image with no raw data"))
	}

	hdr := img.Header()
	nelmts := 1
	for _, dim := range hdr.Axes() {
		nelmts *= dim
	}

	if len(hdr.Axes()) <= 0 {
		nelmts = 0
	}

	//rv := reflect.Indirect(reflect.ValueOf(ptr))
	rv := reflect.ValueOf(ptr).Elem()
	rt := rv.Type()

	if rt.Kind() != reflect.Slice && rt.Kind() != reflect.Array {
		return fmt.Errorf("fitsio: invalid type (%v). expected array or slice", rt.Kind())
	}

	pixsz := hdr.Bitpix() / 8
	if pixsz < 0 {
		pixsz = -pixsz
	}
	otype := rt.Elem()
	if int(otype.Size()) != pixsz {
		return fmt.Errorf(
			"fitsio: element-size do not match. bitpix=%d. elmt-size=%d (conversion not yet supported)",
			hdr.Bitpix(), otype.Size())
	}

	if rt.Kind() == reflect.Slice {
		rv.SetLen(nelmts)
	}

	var cnv = func(v reflect.Value) reflect.Value {
		return v
	}

	bdec := binary.NewDecoder(bytes.NewBuffer(img.raw))
	bdec.Order = binary.BigEndian

	switch hdr.Bitpix() {
	case 8:
		itype := reflect.TypeOf((*byte)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []byte to %s", rt.Name())
		}
		if itype != otype {
			cnv = func(v reflect.Value) reflect.Value {
				return v.Convert(otype)
			}
		}
		for i := 0; i < nelmts; i++ {
			var v byte
			err = bdec.Decode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 16:
		itype := reflect.TypeOf((*int16)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int16 to %s", rt.Name())
		}
		if itype != otype {
			cnv = func(v reflect.Value) reflect.Value {
				return v.Convert(otype)
			}
		}
		for i := 0; i < nelmts; i++ {
			var v int16
			err = bdec.Decode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 32:
		itype := reflect.TypeOf((*int32)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int32 to %s", rt.Name())
		}
		if itype != otype {
			cnv = func(v reflect.Value) reflect.Value {
				return v.Convert(otype)
			}
		}
		for i := 0; i < nelmts; i++ {
			var v int32
			err = bdec.Decode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 64:
		itype := reflect.TypeOf((*int64)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int64 to %s", rt.Name())
		}
		if itype != otype {
			cnv = func(v reflect.Value) reflect.Value {
				return v.Convert(otype)
			}
		}
		for i := 0; i < nelmts; i++ {
			var v int64
			err = bdec.Decode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case -32:
		itype := reflect.TypeOf((*float32)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []float32 to %s", rt.Name())
		}
		if itype != otype {
			cnv = func(v reflect.Value) reflect.Value {
				return v.Convert(otype)
			}
		}
		for i := 0; i < nelmts; i++ {
			var v float32
			err = bdec.Decode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case -64:
		itype := reflect.TypeOf((*byte)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []float64 to %s", rt.Name())
		}
		if itype != otype {
			cnv = func(v reflect.Value) reflect.Value {
				return v.Convert(otype)
			}
		}
		for i := 0; i < nelmts; i++ {
			var v float64
			err = bdec.Decode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	default:
		return fmt.Errorf("invalid image type [bitpix=%d]", hdr.Bitpix())
	}

	return err
}

// Write writes the given image data to the HDU
func (img *imageHDU) Write(data interface{}) error {
	var err error
	rv := reflect.ValueOf(data).Elem()
	if !rv.CanAddr() {
		return fmt.Errorf("fitsio: %T is not addressable", data)
	}

	hdr := img.Header()
	naxes := len(hdr.Axes())
	if naxes == 0 {
		return nil
	}
	nelmts := 1
	for _, dim := range hdr.Axes() {
		nelmts *= dim
	}

	pixsz := hdr.Bitpix() / 8
	if pixsz < 0 {
		pixsz = -pixsz
	}

	img.raw = make([]byte, pixsz*nelmts)
	buf := &sectionWriter{
		buf: img.raw,
		beg: 0,
	}
	enc := binary.NewEncoder(buf)
	enc.Order = binary.BigEndian

	switch data := rv.Interface().(type) {
	case []byte:
		if hdr.Bitpix() != 8 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	case []int8:
		if hdr.Bitpix() != 8 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	case []int16:
		if hdr.Bitpix() != 16 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	case []int32:
		if hdr.Bitpix() != 32 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	case []int64:
		if hdr.Bitpix() != 64 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	case []float32:
		if hdr.Bitpix() != -32 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	case []float64:
		if hdr.Bitpix() != -64 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			err = enc.Encode(&v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
		}

	default:
		return fmt.Errorf("fitsio: invalid image type (%T)", rv.Interface())
	}

	//img.raw = buf.Bytes()
	return err
}

// freeze freezes an Image before writing, finalizing header values.
func (img *imageHDU) freeze() error {
	var err error
	card := img.Header().Get("XTENSION")
	if card != nil {
		return err
	}

	err = img.Header().prepend(Card{
		Name:    "XTENSION",
		Value:   "IMAGE   ",
		Comment: "IMAGE extension",
	})
	if err != nil {
		return err
	}

	return err
}
