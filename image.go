// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
	"fmt"
	"image"
	"reflect"

	"github.com/astrogo/fitsio/fltimg"
	"github.com/gonuts/binary"
)

// Image represents a FITS image
type Image interface {
	HDU
	Read(ptr interface{}) error
	Write(ptr interface{}) error
	Raw() []byte
	Image() image.Image

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

	r := bytes.NewReader(img.raw)
	switch hdr.Bitpix() {
	case 8:
		itype := reflect.TypeOf((*byte)(nil)).Elem()
		if itype == otype {
			slice := rv.Interface().([]byte)
			for i := 0; i < nelmts; i++ {
				err = readByte(r, &slice[i])
				if err != nil {
					return fmt.Errorf("fitsio: %v", err)
				}
			}
			return nil
		}

		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []byte to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}

		for i := 0; i < nelmts; i++ {
			var v byte
			err = readByte(r, &v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 16:
		itype := reflect.TypeOf((*int16)(nil)).Elem()
		if itype == otype {
			slice := rv.Interface().([]int16)
			for i := 0; i < nelmts; i++ {
				err = readI16(r, &slice[i])
				if err != nil {
					return fmt.Errorf("fitsio: %v", err)
				}
			}
			return nil
		}

		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int16 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v int16
			err = readI16(r, &v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 32:
		itype := reflect.TypeOf((*int32)(nil)).Elem()
		if itype == otype {
			slice := rv.Interface().([]int32)
			for i := 0; i < nelmts; i++ {
				err = readI32(r, &slice[i])
				if err != nil {
					return fmt.Errorf("fitsio: %v", err)
				}
			}
			return nil
		}

		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int32 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v int32
			err = readI32(r, &v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 64:
		itype := reflect.TypeOf((*int64)(nil)).Elem()
		if itype == otype {
			slice := rv.Interface().([]int64)
			for i := 0; i < nelmts; i++ {
				err = readI64(r, &slice[i])
				if err != nil {
					return fmt.Errorf("fitsio: %v", err)
				}
			}
			return nil
		}

		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int64 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v int64
			err = readI64(r, &v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case -32:
		itype := reflect.TypeOf((*float32)(nil)).Elem()
		if itype == otype {
			slice := rv.Interface().([]float32)
			for i := 0; i < nelmts; i++ {
				err = readF32(r, &slice[i])
				if err != nil {
					return fmt.Errorf("fitsio: %v", err)
				}
			}
			return nil
		}

		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []float32 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v float32
			err = readF32(r, &v)
			if err != nil {
				return fmt.Errorf("fitsio: %v", err)
			}
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case -64:
		itype := reflect.TypeOf((*float64)(nil)).Elem()
		if itype == otype {
			slice := rv.Interface().([]float64)
			for i := 0; i < nelmts; i++ {
				err = readF64(r, &slice[i])
				if err != nil {
					return fmt.Errorf("fitsio: %v", err)
				}
			}
			return nil
		}

		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []float64 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v float64
			err = readF64(r, &v)
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

// Image returns an image.Image value.
func (img *imageHDU) Image() image.Image {

	// Getting the HDU bitpix and axes.
	header := img.Header()
	bitpix := header.Bitpix()
	axes := header.Axes()
	raw := img.Raw()

	if len(axes) < 2 {
		return nil
	}

	// Image width and height.
	w := axes[0]
	h := axes[1]

	switch {
	case w <= 0:
		return nil
	case h <= 0:
		return nil
	}

	rect := image.Rect(0, 0, w, h)

	switch bitpix {
	case 8:
		img := &image.Gray{
			Pix:    raw,
			Stride: 1 * w,
			Rect:   rect,
		}
		return img

	case 16:
		img := &image.Gray16{
			Pix:    raw,
			Stride: 2 * w,
			Rect:   rect,
		}
		return img

	case 32:
		img := &image.RGBA{
			Pix:    raw,
			Stride: 4 * w,
			Rect:   rect,
		}
		return img

	case 64:
		img := &image.RGBA64{
			Pix:    raw,
			Stride: 8 * w,
			Rect:   rect,
		}
		return img

	case -32:
		img := fltimg.NewGray32(rect, raw)
		return img

	case -64:
		img := fltimg.NewGray64(rect, raw)
		return img

	default:
		panic(fmt.Errorf("fitsio: image with unknown BITPIX value of %d", bitpix))
	}

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
