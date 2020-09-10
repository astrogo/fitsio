// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
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

	for ii, tc := range []struct {
		name    string
		version int
		cards   []Card
		bitpix  int
		axes    []int
		image   interface{}
	}{
		{
			name:    "i8",
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
			name:    "i16",
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
			name:    "i32",
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
			name:    "i64",
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
			name:    "u8",
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
			image: []uint8{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "u16",
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
			image: []uint16{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "u32",
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
			image: []uint32{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "u64",
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
			image: []uint64{
				0, 1, 2, 3,
				4, 5, 6, 7,
				8, 9, 0, 1,
			},
		},
		{
			name:    "f32",
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
			name:    "f64",
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
		fname := fmt.Sprintf("%03d_%s.fits", ii, tc.name)
		t.Run(fname, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				func(i int) {
					var (
						f   *File
						w   *os.File
						r   *os.File
						err error
						hdu HDU
					)

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

						img := NewImage(tc.bitpix, tc.axes)
						defer img.Close()

						err = img.Header().Append(tc.cards...)
						if err != nil {
							t.Fatalf("error appending cards: %v", err)
						}
						hdu = img

						err = img.Write(&tc.image)
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

						data := reflect.New(reflect.TypeOf(tc.image)).Elem()
						data.Set(reflect.MakeSlice(data.Type(), 0, nelmts))
						err = img.Read(data.Addr().Interface())
						if err != nil {
							t.Fatalf("error reading image: %v", err)
						}

						if got, want := data.Interface(), tc.image; !reflect.DeepEqual(got, want) {
							t.Fatalf("expected image:\ngot= %#v\nwant=%#v", got, want)
						}
					}

					hdr := hdu.Header()
					if hdr.bitpix != tc.bitpix {
						t.Fatalf("expected BITPIX=%v. got %v", tc.bitpix, hdr.bitpix)
					}

					if !reflect.DeepEqual(hdr.Axes(), tc.axes) {
						t.Fatalf("expected AXES==%v. got %v (i=%v)", tc.axes, hdr.Axes(), i)
					}

					name := hdu.Name()
					if name != "primary hdu" {
						t.Fatalf("expected EXTNAME==%q. got %q", "primary hdu", name)
					}

					vers := hdu.Version()
					if vers != tc.version {
						t.Fatalf("expected EXTVER==%v. got %v", tc.version, vers)
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

					for _, ref := range tc.cards {
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
		})
	}
}

func TestImageCubeRoundTrip(t *testing.T) {
	strides := []int{2, 3, 6} // non repetitive, aperiodic strides

	data := make([]int16, 2*3*6)
	for i := 0; i < len(data); i++ {
		data[i] = int16(i - 32768)
	}
	var b bytes.Buffer
	fits, err := Create(&b)
	if err != nil {
		t.Fatalf("could not create FITS file: %+v", err)
	}
	im := NewImage(16, strides)
	cards := []Card{
		{Name: "BZERO", Value: 32768},
		{Name: "BSCALE", Value: 1.},
	}
	im.Header().Append(cards...)

	err = im.Write(data)
	if err != nil {
		t.Fatalf("could not write data to image HDU: %+v", err)
	}

	err = fits.Write(im)
	if err != nil {
		t.Fatalf("could not write image to FITS file: %+v", err)
	}

	err = im.Close()
	if err != nil {
		t.Fatalf("could not close image HDU: %+v", err)
	}

	err = fits.Close()
	if err != nil {
		t.Fatalf("could not close FITS file: %+v", err)
	}

	// now read back
	fits, err = Open(&b)
	if err != nil {
		t.Fatalf("could not open FITS file: %+v", err)
	}
	defer fits.Close()

	hdu0 := fits.HDU(0)
	// we are not testing anything but NAXIS<N> and the data here
	img := hdu0.(Image)
	ax := hdu0.Header().Axes()
	for i := range ax {
		if ax[i] != strides[i] {
			t.Errorf("NAXIS %d value of %d did not match expectation of %d", i, ax[i], strides[i])
		}
	}
	readback := make([]int16, len(data))
	err = img.Read(&readback)
	for i := 0; i < len(data); i++ {
		if data[i] != readback[i] {
			t.Errorf("pixel %d value of %d did not match expectation %d", i, readback[i], data[i])
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
		t.Run(fmt.Sprintf("bitpix=%d", test.bitpix), func(t *testing.T) {
			hdu := NewImage(test.bitpix, []int{w, h})
			switch test.bitpix {
			case 8:
				img := test.want.(*image.Gray)
				err := hdu.Write(&img.Pix)
				if err != nil {
					t.Fatalf("image #%d: error writing raw pixels: %v\n", i, err)
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
					t.Fatalf("image #%d: error writing raw pixels: %v\n", i, err)
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
					t.Fatalf("image #%d: error writing raw pixels: %v\n", i, err)
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
					t.Fatalf("image #%d: error writing raw pixels: %v\n", i, err)
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
					t.Fatalf("image #%d: error writing raw pixels: %v\n", i, err)
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
					t.Fatalf("image #%d: error writing raw pixels: %v\n", i, err)
				}
			default:
				t.Fatalf("image #%d: invalid bitpix=%d", i, test.bitpix)
			}
			if !reflect.DeepEqual(hdu.Image(), test.want) {
				t.Fatalf("image #%d:\n got: %v\nwant: %v\n", i, hdu.Image(), test.want)
			}
		})
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
		t.Run(fmt.Sprintf("bitpix=%d-axes=%d", test.bitpix, test.axes), func(t *testing.T) {
			hdu := NewImage(test.bitpix, test.axes)
			img := hdu.Image()
			if img != test.want {
				t.Fatalf("image #%d:\n got: %v\nwant: %v\n", i, img, test.want)
			}
		})
	}

	t.Run("invalid-axes", func(t *testing.T) {
		defer func() {
			e := recover()
			if e == nil {
				t.Fatalf("error: expected a BITPIX-related panic")
			}
			if got, want := e.(error).Error(), "fitsio: image with unknown BITPIX value of 0"; got != want {
				t.Fatalf("invalid panic error:\ngot= %s\nwant=%s", got, want)
			}
		}()

		hdu := NewImage(0, []int{1, 1})
		_ = hdu.Image()
	})
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

func BenchmarkImageWriteI8_10(b *testing.B)     { benchImageWrite(b, 8, 10) }
func BenchmarkImageWriteI8_100(b *testing.B)    { benchImageWrite(b, 8, 100) }
func BenchmarkImageWriteI8_1000(b *testing.B)   { benchImageWrite(b, 8, 1000) }
func BenchmarkImageWriteI8_10000(b *testing.B)  { benchImageWrite(b, 8, 10000) }
func BenchmarkImageWriteI8_100000(b *testing.B) { benchImageWrite(b, 8, 100000) }

func BenchmarkImageWriteI16_10(b *testing.B)     { benchImageWrite(b, 16, 10) }
func BenchmarkImageWriteI16_100(b *testing.B)    { benchImageWrite(b, 16, 100) }
func BenchmarkImageWriteI16_1000(b *testing.B)   { benchImageWrite(b, 16, 1000) }
func BenchmarkImageWriteI16_10000(b *testing.B)  { benchImageWrite(b, 16, 10000) }
func BenchmarkImageWriteI16_100000(b *testing.B) { benchImageWrite(b, 16, 100000) }

func BenchmarkImageWriteI32_10(b *testing.B)     { benchImageWrite(b, 32, 10) }
func BenchmarkImageWriteI32_100(b *testing.B)    { benchImageWrite(b, 32, 100) }
func BenchmarkImageWriteI32_1000(b *testing.B)   { benchImageWrite(b, 32, 1000) }
func BenchmarkImageWriteI32_10000(b *testing.B)  { benchImageWrite(b, 32, 10000) }
func BenchmarkImageWriteI32_100000(b *testing.B) { benchImageWrite(b, 32, 100000) }

func BenchmarkImageWriteI64_10(b *testing.B)     { benchImageWrite(b, 64, 10) }
func BenchmarkImageWriteI64_100(b *testing.B)    { benchImageWrite(b, 64, 100) }
func BenchmarkImageWriteI64_1000(b *testing.B)   { benchImageWrite(b, 64, 1000) }
func BenchmarkImageWriteI64_10000(b *testing.B)  { benchImageWrite(b, 64, 10000) }
func BenchmarkImageWriteI64_100000(b *testing.B) { benchImageWrite(b, 64, 100000) }

func BenchmarkImageWriteF32_10(b *testing.B)     { benchImageWrite(b, -32, 10) }
func BenchmarkImageWriteF32_100(b *testing.B)    { benchImageWrite(b, -32, 100) }
func BenchmarkImageWriteF32_1000(b *testing.B)   { benchImageWrite(b, -32, 1000) }
func BenchmarkImageWriteF32_10000(b *testing.B)  { benchImageWrite(b, -32, 10000) }
func BenchmarkImageWriteF32_100000(b *testing.B) { benchImageWrite(b, -32, 100000) }

func BenchmarkImageWriteF64_10(b *testing.B)     { benchImageWrite(b, -64, 10) }
func BenchmarkImageWriteF64_100(b *testing.B)    { benchImageWrite(b, -64, 100) }
func BenchmarkImageWriteF64_1000(b *testing.B)   { benchImageWrite(b, -64, 1000) }
func BenchmarkImageWriteF64_10000(b *testing.B)  { benchImageWrite(b, -64, 10000) }
func BenchmarkImageWriteF64_100000(b *testing.B) { benchImageWrite(b, -64, 100000) }

func benchImageWrite(b *testing.B, bitpix int, n int) {
	w, err := ioutil.TempFile("", "go-test-fitsio-")
	if err != nil {
		b.Fatal(err)
	}
	defer w.Close()
	defer os.RemoveAll(w.Name())

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

	b.ReportAllocs()
	b.ResetTimer()

	var img *imageHDU
	for i := 0; i < b.N; i++ {
		img = NewImage(bitpix, axes)
		err = img.Write(ptr)
		if err != nil {
			b.Fatal(err)
		}
		img.Close()
	}

	err = f.Write(img)
	if err != nil {
		b.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		b.Fatal(err)
	}

	return
}

func TestPanicWhenNAXISTooLarge(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Fatalf("expected a panic on too many axes")
		}
		if got, want := e.(error).Error(), "fitsio: too many axes (got=1000 > 999)"; got != want {
			t.Fatalf("invalid panic error\ngot= %s\nwant=%s", got, want)
		}
	}()
	ax := make([]int, 1000)
	_ = NewImage(32, ax)
}
