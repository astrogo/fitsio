// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/astrogo/fitsio/fltimg"
)

func TestImageRW(t *testing.T) {
	curdir, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Chdir(curdir)

	workdir, err := ioutil.TempDir("", "go-fitsio-test-")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.RemoveAll(workdir)

	err = os.Chdir(workdir)
	if err != nil {
		t.Fatalf(err.Error())
	}

	for ii, table := range []struct {
		name    string
		version int
		cards   []Card
		bitpix  int
		axes    []int
		image   interface{}
	}{
		{
			name:    "new.fits",
			version: 2,
			cards: []Card{
				{
					"EXTNAME",
					"primary hdu",
					"the primary HDU",
				},
				{
					"EXTVER",
					2,
					"the primary hdu version",
				},
			},
			bitpix: 8,
			axes:   []int{3, 4},
			image: []int8{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "new.fits",
			version: 2,
			cards: []Card{
				{
					"EXTNAME",
					"primary hdu",
					"the primary HDU",
				},
				{
					"EXTVER",
					2,
					"the primary hdu version",
				},
			},
			bitpix: 16,
			axes:   []int{3, 4},
			image: []int16{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "new.fits",
			version: 2,
			cards: []Card{
				{
					"EXTNAME",
					"primary hdu",
					"the primary HDU",
				},
				{
					"EXTVER",
					2,
					"the primary hdu version",
				},
			},
			bitpix: 32,
			axes:   []int{3, 4},
			image: []int32{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "new.fits",
			version: 2,
			cards: []Card{
				{
					"EXTNAME",
					"primary hdu",
					"the primary HDU",
				},
				{
					"EXTVER",
					2,
					"the primary hdu version",
				},
			},
			bitpix: 64,
			axes:   []int{3, 4},
			image: []int64{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "new.fits",
			version: 2,
			cards: []Card{
				{
					"EXTNAME",
					"primary hdu",
					"the primary HDU",
				},
				{
					"EXTVER",
					2,
					"the primary hdu version",
				},
			},
			bitpix: -32,
			axes:   []int{3, 4},
			image: []float32{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "new.fits",
			version: 2,
			cards: []Card{
				{
					"EXTNAME",
					"primary hdu",
					"the primary HDU",
				},
				{
					"EXTVER",
					2,
					"the primary hdu version",
				},
			},
			bitpix: -64,
			axes:   []int{3, 4},
			image: []float64{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
	} {
		fname := fmt.Sprintf("%03d_%s", ii, table.name)
		for i := 0; i < 2; i++ {
			func(i int) {
				var f *File
				var w *os.File
				var r *os.File
				var err error
				var hdu HDU

				switch i {

				case 0: // create
					//fmt.Printf("========= create [%s]....\n", fname)
					w, err = os.Create(fname)
					if err != nil {
						t.Fatalf("error creating new file [%v]: %v", fname, err)
					}
					defer w.Close()

					f, err = Create(w)
					if err != nil {
						t.Fatalf("error creating new file [%v]: %v", fname, err)
					}
					defer f.Close()

					img := NewImage(table.bitpix, table.axes)
					defer img.Close()

					err = img.Header().Append(table.cards...)
					if err != nil {
						t.Fatalf("error appending cards: %v", err)
					}
					hdu = img

					err = img.Write(&table.image)
					if err != nil {
						t.Fatalf("error writing image: %v", err)
					}

					err = f.Write(img)
					if err != nil {
						t.Fatalf("error writing image: %v", err)
					}

				case 1: // read
					//fmt.Printf("========= read [%s]....\n", fname)
					r, err = os.Open(fname)
					if err != nil {
						t.Fatalf("error opening file [%v]: %v", fname, err)
					}
					defer r.Close()
					f, err = Open(r)
					if err != nil {
						t.Fatalf("error opening file [%v]: %v", fname, err)
					}
					defer f.Close()

					hdu = f.HDU(0)
					hdr := hdu.Header()
					img := hdu.(Image)
					nelmts := 1
					for _, axe := range hdr.Axes() {
						nelmts *= int(axe)
					}

					var data interface{}
					switch hdr.Bitpix() {
					case 8:
						v := make([]int8, 0, nelmts)
						err = img.Read(&v)
						data = v

					case 16:
						v := make([]int16, 0, nelmts)
						err = img.Read(&v)
						data = v

					case 32:
						v := make([]int32, 0, nelmts)
						err = img.Read(&v)
						data = v

					case 64:
						v := make([]int64, 0, nelmts)
						err = img.Read(&v)
						data = v

					case -32:
						v := make([]float32, 0, nelmts)
						err = img.Read(&v)
						data = v

					case -64:
						v := make([]float64, 0, nelmts)
						err = img.Read(&v)
						data = v
					}

					if err != nil {
						t.Fatalf("error reading image: %v", err)
					}

					if !reflect.DeepEqual(data, table.image) {
						t.Fatalf("expected image:\nref=%v\ngot=%v", table.image, data)
					}
				}

				hdr := hdu.Header()
				if hdr.bitpix != table.bitpix {
					t.Fatalf("expected BITPIX=%v. got %v", table.bitpix, hdr.bitpix)
				}

				if !reflect.DeepEqual(hdr.Axes(), table.axes) {
					t.Fatalf("expected AXES==%v. got %v (i=%v)", table.axes, hdr.Axes(), i)
				}

				name := hdu.Name()
				if name != "primary hdu" {
					t.Fatalf("expected EXTNAME==%q. got %q", "primary hdu", name)
				}

				vers := hdu.Version()
				if vers != table.version {
					t.Fatalf("expected EXTVER==%v. got %v", table.version, vers)
				}

				card := hdr.Get("EXTNAME")
				if card == nil {
					t.Fatalf("error retrieving card [EXTNAME]")
				}
				if card.Comment != "the primary HDU" {
					t.Fatalf("expected EXTNAME.Comment==%q. got %q", "the primary HDU", card.Comment)
				}

				card = hdr.Get("EXTVER")
				if card == nil {
					t.Fatalf("error retrieving card [EXTVER]")
				}
				if card.Comment != "the primary hdu version" {
					t.Fatalf("expected EXTVER.Comment==%q. got %q", "the primary hdu version", card.Comment)

				}

				for _, ref := range table.cards {
					card := hdr.Get(ref.Name)
					if card == nil {
						t.Fatalf("error retrieving card [%v]", ref.Name)
					}
					rv := reflect.ValueOf(ref.Value)
					var val interface{}
					switch rv.Type().Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						val = int(rv.Int())
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						val = int(rv.Uint())
					case reflect.Float32, reflect.Float64:
						val = rv.Float()
					case reflect.Complex64, reflect.Complex128:
						val = rv.Complex()
					case reflect.String:
						val = ref.Value.(string)
					}
					if !reflect.DeepEqual(card.Value, val) {
						t.Fatalf(
							"card %q. expected [%v](%T). got [%v](%T)",
							ref.Name,
							val, val,
							card.Value, card.Value,
						)
					}
					if card.Comment != ref.Comment {
						t.Fatalf("card %q. comment differ. expected %q. got %q", ref.Name, ref.Comment, card.Comment)
					}
				}

				card = hdr.Get("NOT THERE")
				if card != nil {
					t.Fatalf("expected no card. got [%v]", card)
				}
			}(i)
		}
	}
}

func TestImageImage(t *testing.T) {
	const (
		w = 20
		h = 20
	)
	var (
		rect = image.Rect(0, 0, w, h)
	)
	set := func(img draw.Image) {
		img.Set(0, 0, color.RGBA{255, 0, 0, 255})
		img.Set(10, 10, color.RGBA{0, 255, 0, 255})
		img.Set(15, 15, color.RGBA{0, 0, 255, 255})
	}

	for i, test := range []struct {
		want   image.Image
		bitpix int
	}{
		{
			want: func() image.Image {
				img := image.NewGray(rect)
				set(img)
				return img
			}(),
			bitpix: 8,
		},
		{
			want: func() image.Image {
				img := image.NewGray16(rect)
				set(img)
				return img
			}(),
			bitpix: 16,
		},
		{
			want: func() image.Image {
				img := image.NewRGBA(rect)
				set(img)
				return img
			}(),
			bitpix: 32,
		},
		{
			want: func() image.Image {
				img := image.NewRGBA64(rect)
				set(img)
				return img
			}(),
			bitpix: 64,
		},
		{
			want: func() image.Image {
				pix := image.NewRGBA(rect)
				set(pix)
				img := fltimg.NewGray32(rect, pix.Pix)
				return img
			}(),
			bitpix: -32,
		},
		{
			want: func() image.Image {
				pix := image.NewRGBA64(rect)
				set(pix)
				img := fltimg.NewGray64(rect, pix.Pix)
				return img
			}(),
			bitpix: -64,
		},
	} {
		hdu := NewImage(test.bitpix, []int{w, h})
		switch test.bitpix {
		case 8:
			img := test.want.(*image.Gray)
			err := hdu.Write(&img.Pix)
			if err != nil {
				t.Errorf("image #%d: error writing raw pixels: %v\n", i, err)
				continue
			}
		case 16:
			img := test.want.(*image.Gray16)
			pix := make([]int16, len(img.Pix)/2)
			for i := 0; i < len(img.Pix); i += 2 {
				buf := img.Pix[i : i+2]
				pix[i/2] = int16(uint16(buf[1]) | uint16(buf[0])<<8)
			}
			err := hdu.Write(&pix)
			if err != nil {
				t.Errorf("image #%d: error writing raw pixels: %v\n", i, err)
				continue
			}
		case 32:
			img := test.want.(*image.RGBA)
			pix := make([]int32, len(img.Pix)/4)
			for i := 0; i < len(img.Pix); i += 4 {
				buf := img.Pix[i : i+4]
				pix[i/4] = int32(uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24)
			}
			err := hdu.Write(&pix)
			if err != nil {
				t.Errorf("image #%d: error writing raw pixels: %v\n", i, err)
				continue
			}
		case 64:
			img := test.want.(*image.RGBA64)
			pix := make([]int64, len(img.Pix)/8)
			for i := 0; i < len(img.Pix); i += 8 {
				buf := img.Pix[i : i+8]
				pix[i/8] = int64(uint64(buf[7]) | uint64(buf[6])<<8 | uint64(buf[5])<<16 | uint64(buf[4])<<24 |
					uint64(buf[3])<<32 | uint64(buf[2])<<40 | uint64(buf[1])<<48 | uint64(buf[0])<<56)
			}
			err := hdu.Write(&pix)
			if err != nil {
				t.Errorf("image #%d: error writing raw pixels: %v\n", i, err)
				continue
			}
		case -32:
			img := test.want.(*fltimg.Gray32)
			pix := make([]float32, len(img.Pix)/4)
			for i := 0; i < len(img.Pix); i += 4 {
				buf := img.Pix[i : i+4]
				pix[i/4] = math.Float32frombits(uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24)
			}
			err := hdu.Write(&pix)
			if err != nil {
				t.Errorf("image #%d: error writing raw pixels: %v\n", i, err)
				continue
			}
		case -64:
			img := test.want.(*fltimg.Gray64)
			pix := make([]float64, len(img.Pix)/8)
			for i := 0; i < len(img.Pix); i += 8 {
				buf := img.Pix[i : i+8]
				pix[i/8] = math.Float64frombits(uint64(buf[7]) | uint64(buf[6])<<8 | uint64(buf[5])<<16 | uint64(buf[4])<<24 |
					uint64(buf[3])<<32 | uint64(buf[2])<<40 | uint64(buf[1])<<48 | uint64(buf[0])<<56)
			}
			err := hdu.Write(&pix)
			if err != nil {
				t.Errorf("image #%d: error writing raw pixels: %v\n", i, err)
				continue
			}
		default:
			t.Errorf("image #%d: invalid bitpix=%d", i, test.bitpix)
			continue
		}
		if !reflect.DeepEqual(hdu.Image(), test.want) {
			t.Errorf("image #%d:\n got: %v\nwant: %v\n", i, hdu.Image(), test.want)
			continue
		}
	}

	for i, test := range []struct {
		bitpix int
		axes   []int
		want   image.Image
	}{
		{
			bitpix: 0,
			axes:   nil,
		},
		{
			bitpix: 0,
			axes:   []int{},
		},
		{
			bitpix: 0,
			axes:   []int{1},
		},
		{
			bitpix: 0,
			axes:   []int{0, 0},
		},
		{
			bitpix: 8,
			axes:   nil,
		},
		{
			bitpix: 8,
			axes:   []int{},
		},
		{
			bitpix: 8,
			axes:   []int{1},
		},
		{
			bitpix: 8,
			axes:   []int{0, 0},
		},
	} {
		hdu := NewImage(test.bitpix, test.axes)
		img := hdu.Image()
		if img != test.want {
			t.Errorf("image #%d:\n got: %v\nwant: %v\n", i, img, test.want)
			continue
		}
	}

	func() {
		defer func() {
			if e := recover(); e == nil {
				t.Errorf("error: expected a BITPIX-related panic")
			}
		}()

		hdu := NewImage(0, []int{1, 1})
		_ = hdu.Image()
	}()
}

func BenchmarkImageReadI8_10(b *testing.B)     { benchImageRead(b, 8, 10) }
func BenchmarkImageReadI8_100(b *testing.B)    { benchImageRead(b, 8, 100) }
func BenchmarkImageReadI8_1000(b *testing.B)   { benchImageRead(b, 8, 1000) }
func BenchmarkImageReadI8_10000(b *testing.B)  { benchImageRead(b, 8, 10000) }
func BenchmarkImageReadI8_100000(b *testing.B) { benchImageRead(b, 8, 100000) }

func BenchmarkImageReadI16_10(b *testing.B)     { benchImageRead(b, 16, 10) }
func BenchmarkImageReadI16_100(b *testing.B)    { benchImageRead(b, 16, 100) }
func BenchmarkImageReadI16_1000(b *testing.B)   { benchImageRead(b, 16, 1000) }
func BenchmarkImageReadI16_10000(b *testing.B)  { benchImageRead(b, 16, 10000) }
func BenchmarkImageReadI16_100000(b *testing.B) { benchImageRead(b, 16, 100000) }

func BenchmarkImageReadI32_10(b *testing.B)     { benchImageRead(b, 32, 10) }
func BenchmarkImageReadI32_100(b *testing.B)    { benchImageRead(b, 32, 100) }
func BenchmarkImageReadI32_1000(b *testing.B)   { benchImageRead(b, 32, 1000) }
func BenchmarkImageReadI32_10000(b *testing.B)  { benchImageRead(b, 32, 10000) }
func BenchmarkImageReadI32_100000(b *testing.B) { benchImageRead(b, 32, 100000) }

func BenchmarkImageReadI64_10(b *testing.B)     { benchImageRead(b, 64, 10) }
func BenchmarkImageReadI64_100(b *testing.B)    { benchImageRead(b, 64, 100) }
func BenchmarkImageReadI64_1000(b *testing.B)   { benchImageRead(b, 64, 1000) }
func BenchmarkImageReadI64_10000(b *testing.B)  { benchImageRead(b, 64, 10000) }
func BenchmarkImageReadI64_100000(b *testing.B) { benchImageRead(b, 64, 100000) }

func BenchmarkImageReadF32_10(b *testing.B)     { benchImageRead(b, -32, 10) }
func BenchmarkImageReadF32_100(b *testing.B)    { benchImageRead(b, -32, 100) }
func BenchmarkImageReadF32_1000(b *testing.B)   { benchImageRead(b, -32, 1000) }
func BenchmarkImageReadF32_10000(b *testing.B)  { benchImageRead(b, -32, 10000) }
func BenchmarkImageReadF32_100000(b *testing.B) { benchImageRead(b, -32, 100000) }

func BenchmarkImageReadF64_10(b *testing.B)     { benchImageRead(b, -64, 10) }
func BenchmarkImageReadF64_100(b *testing.B)    { benchImageRead(b, -64, 100) }
func BenchmarkImageReadF64_1000(b *testing.B)   { benchImageRead(b, -64, 1000) }
func BenchmarkImageReadF64_10000(b *testing.B)  { benchImageRead(b, -64, 10000) }
func BenchmarkImageReadF64_100000(b *testing.B) { benchImageRead(b, -64, 100000) }

func benchImageRead(b *testing.B, bitpix int, n int) {
	fname := genImage(b, bitpix, n)
	defer os.RemoveAll(fname)

	r, err := os.Open(fname)
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()

	f, err := Open(r)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	hdu := f.HDU(0)
	img := hdu.(Image)
	var ptr interface{}
	switch bitpix {
	case 8:
		raw := make([]int8, 0, n)
		ptr = &raw
	case 16:
		raw := make([]int16, 0, n)
		ptr = &raw
	case 32:
		raw := make([]int32, 0, n)
		ptr = &raw
	case 64:
		raw := make([]int64, 0, n)
		ptr = &raw
	case -32:
		raw := make([]float32, 0, n)
		ptr = &raw
	case -64:
		raw := make([]float64, 0, n)
		ptr = &raw
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = img.Read(ptr)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func genImage(b *testing.B, bitpix int, n int) string {
	w, err := ioutil.TempFile("", "go-test-fitsio-")
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()

	f, err := Create(w)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	var ptr interface{}
	axes := []int{n}
	switch bitpix {
	case 8:
		data := make([]int8, n)
		for i := range data {
			data[i] = int8(i)
		}
		ptr = &data

	case 16:
		data := make([]int16, n)
		for i := range data {
			data[i] = int16(i)
		}
		ptr = &data

	case 32:
		data := make([]int32, n)
		for i := range data {
			data[i] = int32(i)
		}
		ptr = &data

	case 64:
		data := make([]int64, n)
		for i := range data {
			data[i] = int64(i)
		}
		ptr = &data

	case -32:
		data := make([]float32, n)
		for i := range data {
			data[i] = float32(i)
		}
		ptr = &data

	case -64:
		data := make([]float64, n)
		for i := range data {
			data[i] = float64(i)
		}
		ptr = &data
	default:
		panic(fmt.Errorf("invalid bitpix=%d", bitpix))
	}

	img := NewImage(bitpix, axes)
	defer img.Close()
	err = img.Write(ptr)
	if err != nil {
		b.Fatal(err)
	}

	err = f.Write(img)
	if err != nil {
		b.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		b.Fatal(err)
	}

	return w.Name()
}
