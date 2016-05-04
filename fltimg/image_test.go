// Copyright 2016 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fltimg

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"math/rand"
	"testing"
)

func TestGray32Model(t *testing.T) {
	for i, c := range []color.Color{
		color.RGBAModel.Convert(f32Gray(0)),
		color.RGBAModel.Convert(f32Gray(0.25)),
		color.RGBAModel.Convert(f32Gray(0.5)),
		color.RGBAModel.Convert(f32Gray(0.99)),
		color.RGBAModel.Convert(f32Gray(1)),

		color.RGBAModel.Convert(f64Gray(0)),
		color.RGBAModel.Convert(f64Gray(0.25)),
		color.RGBAModel.Convert(f64Gray(0.5)),
		color.RGBAModel.Convert(f64Gray(0.99)),
		color.RGBAModel.Convert(f64Gray(1)),

		f32Gray(0),
		f32Gray(0.5),
		f32Gray(1),
	} {
		r, g, b, a := Gray32Model.Convert(c).RGBA()
		got := [4]uint32{r, g, b, a}
		r, g, b, a = c.RGBA()
		want := [4]uint32{r, g, b, a}

		if got != want {
			t.Errorf("test[%d]: got=%v want=%v\n", i, got, want)
		}
	}
}

func TestGray64Model(t *testing.T) {
	for i, c := range []color.Color{
		color.RGBAModel.Convert(f32Gray(0)),
		color.RGBAModel.Convert(f32Gray(0.25)),
		color.RGBAModel.Convert(f32Gray(0.5)),
		color.RGBAModel.Convert(f32Gray(0.99)),
		color.RGBAModel.Convert(f32Gray(1)),

		color.RGBAModel.Convert(f64Gray(0)),
		color.RGBAModel.Convert(f64Gray(0.25)),
		color.RGBAModel.Convert(f64Gray(0.5)),
		color.RGBAModel.Convert(f64Gray(0.99)),
		color.RGBAModel.Convert(f64Gray(1)),

		f64Gray(0),
		f64Gray(0.5),
		f64Gray(1),
	} {
		r, g, b, a := Gray64Model.Convert(c).RGBA()
		got := [4]uint32{r, g, b, a}
		r, g, b, a = c.RGBA()
		want := [4]uint32{r, g, b, a}

		if got != want {
			t.Errorf("test[%d]: got=%v want=%v\n", i, got, want)
		}
	}
}

func TestGray32(t *testing.T) {
	rect := image.Rect(0, 0, 50, 50)
	pix := make([]float32, rect.Dx()*rect.Dy())

	for i := range pix {
		v := rand.Float32()
		pix[i] = v
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, pix)
	raw := buf.Bytes()

	img := NewGray32(rect, raw)

	nfail := 0
	for i := 0; i < len(raw); i += 4 {
		want := pix[i/4]
		v1 := img.at(i)
		if v1 != want {
			t.Errorf("img[%d]: got=%v want=%v\n", i, v1, want)
			nfail++
		}

		img.setf(i, want+0.5)
		v2 := img.at(i)
		if v2 != want+0.5 {
			t.Errorf("img[%d]: got=%v want=%v\n", i, v2, want+0.5)
			nfail++
		}
		img.setf(i, want)

		v3 := img.at(i)
		if v3 != want {
			t.Errorf("img[%d]: got=%v want=%v\n", i, v3, want)
			nfail++
		}

		if nfail > 20 {
			t.Fatalf("too many failures")
		}
	}
}

func TestGray64(t *testing.T) {
	rect := image.Rect(0, 0, 50, 50)
	pix := make([]float64, rect.Dx()*rect.Dy())

	for i := range pix {
		v := rand.Float64()
		pix[i] = v
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, pix)
	raw := buf.Bytes()

	img := NewGray64(rect, raw)

	nfail := 0
	for i := 0; i < len(raw); i += 8 {
		want := pix[i/8]
		v1 := img.at(i)
		if v1 != want {
			t.Errorf("img[%d]: got=%v want=%v\n", i, v1, want)
			nfail++
		}

		img.setf(i, want+0.5)
		v2 := img.at(i)
		if v2 != want+0.5 {
			t.Errorf("img[%d]: got=%v want=%v\n", i, v2, want+0.5)
			nfail++
		}
		img.setf(i, want)

		v3 := img.at(i)
		if v3 != want {
			t.Errorf("img[%d]: got=%v want=%v\n", i, v3, want)
			nfail++
		}

		if nfail > 20 {
			t.Fatalf("too many failures")
		}
	}
}
