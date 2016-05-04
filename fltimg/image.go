// Copyright 2016 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// fltimg provides image.Image implementations for the float32- and float64-image encodings of FITS.
package fltimg

import (
	"encoding/binary"
	"image"
	"image/color"
	"math"
)

const (
	gamma = 1 / 2.2
)

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

// Gray32 represents an image.Image encoded in 32b IEEE floating-point values
type Gray32 struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
	Min    float32
	Max    float32
}

// NewGray32 creates a new Gray32 image with the given bounds.
func NewGray32(rect image.Rectangle, pix []byte) *Gray32 {
	w, h := rect.Dx(), rect.Dy()
	switch {
	case pix == nil:
		panic("fltimg: nil pixel buffer")
	case len(pix) != 4*w*h:
		panic("fltimg: inconsistent pixels size")
	}
	img := &Gray32{Pix: pix, Stride: 4 * w, Rect: rect}
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

func (p *Gray32) at(i int) float32 {
	bits := binary.BigEndian.Uint32(p.Pix[i : i+4])
	return math.Float32frombits(bits)
}

func (p *Gray32) setf(i int, v float32) {
	binary.BigEndian.PutUint32(p.Pix[i:i+4], math.Float32bits(v))
}

func (p *Gray32) ColorModel() color.Model { return Gray32Model }
func (p *Gray32) Bounds() image.Rectangle { return p.Rect }
func (p *Gray32) At(x, y int) color.Color {
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

func (p *Gray32) Set(x, y int, c color.Color) {
	i := p.PixOffset(x, y)
	r, _, _, _ := Gray32Model.Convert(c).RGBA()
	v := math.Exp(math.Log(float64(r)/float64(0xffff)) / gamma)
	p.setf(i, float32(v))
}

func (p *Gray32) PixOffset(x, y int) int {
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
	i := uint32(f * 0xffff)
	return i, i, i, 0xffff
}

// Gray64 represents an image.Image encoded in 64b IEEE floating-point values
type Gray64 struct {
	Pix    []uint8
	Stride int
	Rect   image.Rectangle
	Min    float64
	Max    float64
}

// NewGray64 creates a new Gray64 image with the given bounds.
func NewGray64(rect image.Rectangle, pix []byte) *Gray64 {
	w, h := rect.Dx(), rect.Dy()
	switch {
	case pix == nil:
		panic("fltimg: nil pixel buffer")
	case len(pix) != 8*w*h:
		panic("fltimg: inconsistent pixels size")
	}
	img := &Gray64{pix, 8 * w, rect, 0, 0}
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

func (p *Gray64) at(i int) float64 {
	bits := binary.BigEndian.Uint64(p.Pix[i : i+8])
	return math.Float64frombits(bits)
}

func (p *Gray64) setf(i int, v float64) {
	binary.BigEndian.PutUint64(p.Pix[i:i+8], math.Float64bits(v))
}

func (p *Gray64) ColorModel() color.Model { return Gray64Model }
func (p *Gray64) Bounds() image.Rectangle { return p.Rect }
func (p *Gray64) At(x, y int) color.Color {
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

func (p *Gray64) Set(x, y int, c color.Color) {
	i := p.PixOffset(x, y)
	r, _, _, _ := Gray64Model.Convert(c).RGBA()
	v := math.Exp(math.Log(float64(r)/float64(0xffff)) / gamma)
	p.setf(i, v)
}

func (p *Gray64) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*8
}

// Models for the fltimg color types.
var (
	Gray32Model color.Model = color.ModelFunc(gray32Model)
	Gray64Model color.Model = color.ModelFunc(gray64Model)
)

func gray32Model(c color.Color) color.Color {
	if _, ok := c.(f32Gray); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	y := (19595*r + 38470*g + 7471*b + 1<<15) >> 16
	v := math.Exp(math.Log(float64(y)/float64(0xffff)) / gamma)
	return f32Gray(v)
}

func gray64Model(c color.Color) color.Color {
	if _, ok := c.(f64Gray); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	y := (19595*r + 38470*g + 7471*b + 1<<15) >> 16
	v := math.Exp(math.Log(float64(y)/float64(0xffff)) / gamma)
	return f64Gray(v)
}
