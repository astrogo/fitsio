// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"

	"github.com/gonuts/binary"
)

const (
	gamma = 1 / 2.2
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

// Image returns an image.Image value.
func (img *imageHDU) Image() image.Image {

	// Getting the HDU bitpix and axes.
	header := img.Header()
	bitpix := header.Bitpix()
	axes := header.Axes()
	raw := img.Raw()

	// Image width and height.
	w := axes[0]
	h := axes[1]

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
		img := newF32Image(rect, raw)
		return img

	case -64:
		img := newF64Image(rect, raw)
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

type f32Gray float32

func (c f32Gray) RGBA() (r, g, b, a uint32) {
	f := math.Pow(float64(c), gamma)
	switch {
	case f > 1:
		f = 1
	case f < 0:
		f = 0
	}
	i := uint32(f * 0xffff)
	return i, i, i, 0xffff
}

// f32Image represents an Image with a bitpix=-32
type f32Image struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
	Min    float32
	Max    float32
}

// newF32Image creates a new f32Image with the given bounds.
func newF32Image(rect image.Rectangle, pix []byte) *f32Image {
	w, h := rect.Dx(), rect.Dy()
	switch {
	case pix == nil:
		panic("fitsio: nil pixel buffer")
	case len(pix) != 4*w*h:
		panic("fitsio: inconsistent pixels size")
	}
	img := &f32Image{Pix: pix, Stride: 4 * w, Rect: rect}
	min := float32(+math.MaxFloat32)
	max := float32(-math.MaxFloat32)
	for i := 0; i < len(img.Pix); i += 4 {
		v := img.at(i)
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	img.Min = min
	img.Max = max
	return img
}

func (p *f32Image) at(i int) float32 {
	buf := p.Pix[i : i+4]
	bits := uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24
	return math.Float32frombits(bits)
}

func (p *f32Image) ColorModel() color.Model { return color.RGBAModel }
func (p *f32Image) Bounds() image.Rectangle { return p.Rect }
func (p *f32Image) At(x, y int) color.Color {
	i := p.PixOffset(x, y)
	v := p.at(i)
	f := (v - p.Min) / (p.Max - p.Min)
	switch {
	case f < 0:
		f = 0
	case f > 1:
		f = 1
	}
	return f32Gray(f)
}

func (p *f32Image) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*4
}

type f64Gray float64

func (c f64Gray) RGBA() (r, g, b, a uint32) {
	f := math.Pow(float64(c), gamma)
	switch {
	case f > 1:
		f = 1
	case f < 0:
		f = 0
	}
	i := uint32(f * 0xffffffff)
	return i, i, i, 0xffff
}

// f64Image represents an Image with a bitpix=-64
type f64Image struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
	Min    float64
	Max    float64
}

// newF64Image creates a new f64Image with the given bounds.
func newF64Image(rect image.Rectangle, pix []byte) *f64Image {
	w, h := rect.Dx(), rect.Dy()
	switch {
	case pix == nil:
		panic("fitsio: nil pixel buffer")
	case len(pix) != 8*w*h:
		panic("fitsio: inconsistent pixels size")
	}
	img := &f64Image{pix, 8 * w, rect, 0, 0}
	min := +math.MaxFloat64
	max := -math.MaxFloat64
	for i := 0; i < len(img.Pix); i += 8 {
		v := img.at(i)
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	img.Min = min
	img.Max = max
	return img

}

func (p *f64Image) at(i int) float64 {
	buf := p.Pix[i : i+8]
	bits := uint64(buf[7]) | uint64(buf[6])<<8 | uint64(buf[5])<<16 | uint64(buf[4])<<24 |
		uint64(buf[3])<<32 | uint64(buf[2])<<40 | uint64(buf[1])<<48 | uint64(buf[0])<<56
	return math.Float64frombits(bits)
}

func (p *f64Image) ColorModel() color.Model { return color.RGBA64Model }
func (p *f64Image) Bounds() image.Rectangle { return p.Rect }
func (p *f64Image) At(x, y int) color.Color {
	i := p.PixOffset(x, y)
	v := p.at(i)
	f := (1 - (v-p.Min)/(p.Max-p.Min))
	switch {
	case f < 0:
		f = 0
	case f > 1:
		f = 1
	}
	return f64Gray(f)
}

func (p *f64Image) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*8
}
