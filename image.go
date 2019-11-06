// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"reflect"

	"github.com/astrogo/fitsio/fltimg"
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

	r := newReader(img.raw)
	switch hdr.Bitpix() {
	case 8:
		switch slice := rv.Interface().(type) {
		case []int8:
			for i := 0; i < nelmts; i++ {
				r.readI8(&slice[i])
			}
			return nil
		case []byte:
			for i := 0; i < nelmts; i++ {
				r.readByte(&slice[i])
			}
			return nil
		}

		itype := reflect.TypeOf((*byte)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []byte to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}

		for i := 0; i < nelmts; i++ {
			var v byte
			r.readByte(&v)
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 16:
		if otype.Kind() == reflect.Int16 {
			slice := rv.Interface().([]int16)
			for i := 0; i < nelmts; i++ {
				r.readI16(&slice[i])
			}
			return nil
		}

		itype := reflect.TypeOf((*int16)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int16 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v int16
			r.readI16(&v)
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 32:
		if otype.Kind() == reflect.Int32 {
			slice := rv.Interface().([]int32)
			for i := 0; i < nelmts; i++ {
				r.readI32(&slice[i])
			}
			return nil
		}

		itype := reflect.TypeOf((*int32)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int32 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v int32
			r.readI32(&v)
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case 64:
		if otype.Kind() == reflect.Int64 {
			slice := rv.Interface().([]int64)
			for i := 0; i < nelmts; i++ {
				r.readI64(&slice[i])
			}
			return nil
		}

		itype := reflect.TypeOf((*int64)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []int64 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v int64
			r.readI64(&v)
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case -32:
		if otype.Kind() == reflect.Float32 {
			slice := rv.Interface().([]float32)
			for i := 0; i < nelmts; i++ {
				r.readF32(&slice[i])
			}
			return nil
		}

		itype := reflect.TypeOf((*float32)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []float32 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v float32
			r.readF32(&v)
			rv.Index(i).Set(cnv(reflect.ValueOf(v)))
		}

	case -64:
		if otype.Kind() == reflect.Float64 {
			slice := rv.Interface().([]float64)
			for i := 0; i < nelmts; i++ {
				r.readF64(&slice[i])
			}
			return nil
		}

		itype := reflect.TypeOf((*float64)(nil)).Elem()
		if !rt.Elem().ConvertibleTo(itype) {
			return fmt.Errorf("fitsio: can not convert []float64 to %s", rt.Name())
		}
		cnv := func(v reflect.Value) reflect.Value {
			return v.Convert(otype)
		}
		for i := 0; i < nelmts; i++ {
			var v float64
			r.readF64(&v)
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
	rv := reflect.Indirect(reflect.ValueOf(data))

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
	w := newWriter(img.raw)
	switch data := rv.Interface().(type) {
	case []byte:
		if hdr.Bitpix() != 8 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		copy(img.raw, data)

	case []int8:
		if hdr.Bitpix() != 8 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			w.writeI8(v)
		}

	case []int16:
		if hdr.Bitpix() != 16 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			w.writeI16(v)
		}

	case []int32:
		if hdr.Bitpix() != 32 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			w.writeI32(v)
		}

	case []int64:
		if hdr.Bitpix() != 64 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			w.writeI64(v)
		}

	case []float32:
		if hdr.Bitpix() != -32 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			w.writeF32(v)
		}

	case []float64:
		if hdr.Bitpix() != -64 {
			return fmt.Errorf("fitsio: got a %T but bitpix!=%d", data, hdr.Bitpix())
		}
		for _, v := range data {
			w.writeF64(v)
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
	raw := make([]byte, len(img.Raw()))
	copy(raw, img.Raw())

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

	// handle BSCALE/BZERO rescaling header keywords.
	// see: https://heasarc.gsfc.nasa.gov/docs/fcg/standard_dict.html
	if bscale, bzero := header.Get("BSCALE"), header.Get("BZERO"); bscale != nil || bzero != nil {
		scale := 1.0
		zero := 0.0
		if bzero != nil {
			switch v := bzero.Value.(type) {
			case float64:
				zero = v
			case int:
				zero = float64(v)
			default:
				panic("fitsio: handle non-float types for BSCALE and BZERO")
			}
		}
		if bscale != nil {
			switch v := bscale.Value.(type) {
			case float64:
				scale = v
			case int:
				scale = float64(v)
			default:
				panic("fitsio: handle non-float types for BSCALE and BZERO")
			}
		}
		switch bitpix {
		case 8:
			zero := uint8(zero)
			scale := uint8(scale)
			for i, v := range raw {
				raw[i] = zero + scale*v
			}

		case 16:
			zero := uint16(zero)
			scale := uint16(scale)
			for i := 0; i < len(raw); i += 2 {
				v := zero + scale*binary.BigEndian.Uint16(raw[i:i+2])
				binary.BigEndian.PutUint16(raw[i:i+2], v)
			}

		case 32:
			zero := uint32(zero)
			scale := uint32(scale)
			for i := 0; i < len(raw); i += 4 {
				v := zero + scale*binary.BigEndian.Uint32(raw[i:i+4])
				binary.BigEndian.PutUint32(raw[i:i+4], v)
			}

		case 64:
			zero := uint64(zero)
			scale := uint64(scale)
			for i := 0; i < len(raw); i += 8 {
				v := zero + scale*binary.BigEndian.Uint64(raw[i:i+8])
				binary.BigEndian.PutUint64(raw[i:i+8], v)
			}

		case -32:
			zero := float32(zero)
			scale := float32(scale)
			for i := 0; i < len(raw); i += 4 {
				v := zero + scale*math.Float32frombits(binary.BigEndian.Uint32(raw[i:i+4]))
				binary.BigEndian.PutUint32(raw[i:i+4], math.Float32bits(v))
			}

		case -64:
			for i := 0; i < len(raw); i += 8 {
				v := zero + scale*math.Float64frombits(binary.BigEndian.Uint64(raw[i:i+8]))
				binary.BigEndian.PutUint64(raw[i:i+8], math.Float64bits(v))
			}
		}
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
