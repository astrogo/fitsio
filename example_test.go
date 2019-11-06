// Copyright 2019 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio_test

import (
	"fmt"
	"log"
	"os"

	"github.com/astrogo/fitsio"
)

func Example_open() {
	f, err := os.Open("testdata/file-img2-bitpix+08.fits")
	if err != nil {
		log.Fatalf("could not open file: %+v", err)
	}
	defer f.Close()

	fits, err := fitsio.Open(f)
	if err != nil {
		log.Fatalf("could not open FITS file: %+v", err)
	}
	defer fits.Close()

	for i, hdu := range fits.HDUs() {
		fmt.Printf("Header listing for HDU #%d:\n", i)
		hdr := hdu.Header()
		for k := range hdr.Keys() {
			card := hdr.Card(k)
			fmt.Printf(
				"%-8s= %-29s / %s\n",
				card.Name,
				fmt.Sprintf("%v", card.Value),
				card.Comment)
		}
		fmt.Printf("END\n\n")
	}

	// Output:
	// Header listing for HDU #0:
	// SIMPLE  = true                          / primary HDU
	// BITPIX  = 8                             / number of bits per data pixel
	// NAXIS   = 2                             / number of data axes
	// NAXIS1  = 50                            / length of data axis 1
	// NAXIS2  = 50                            / length of data axis 2
	// IMAGENBR= 0                             / image number
	// END
	//
	// Header listing for HDU #1:
	// XTENSION= IMAGE                         / IMAGE extension
	// BITPIX  = 8                             / number of bits per data pixel
	// NAXIS   = 2                             / number of data axes
	// NAXIS1  = 50                            / length of data axis 1
	// NAXIS2  = 50                            / length of data axis 2
	// IMAGENBR= 1                             / image number
	// END
}

func Example_openImage3x4F64() {
	f, err := os.Open("testdata/example.fits")
	if err != nil {
		log.Fatalf("could not open file: %+v", err)
	}
	defer f.Close()

	fits, err := fitsio.Open(f)
	if err != nil {
		log.Fatalf("could not open FITS file: %+v", err)
	}
	defer fits.Close()

	var (
		hdu    = fits.HDU(0)
		hdr    = hdu.Header()
		img    = hdu.(fitsio.Image)
		bitpix = hdr.Bitpix()
		axes   = hdr.Axes()
	)
	fmt.Printf("bitpix: %d\n", bitpix)
	fmt.Printf("axes:   %v\n", axes)
	data := make([]float64, axes[0]*axes[1])
	err = img.Read(&data)
	if err != nil {
		log.Fatalf("could not read image: %+v", err)
	}
	fmt.Printf("data:   %v\n", data)

	// Output:
	// bitpix: -64
	// axes:   [3 4]
	// data:   [0 1 2 3 4 5 6 7 8 9 0 1]
}

func Example_createImage3x4I8() {
	f, err := os.Create("testdata/ex-img-bitpix+8.fits")
	if err != nil {
		log.Fatalf("could not create file: %+v", err)
	}
	defer f.Close()

	fits, err := fitsio.Create(f)
	if err != nil {
		log.Fatalf("could not create FITS file: %+v", err)
	}
	defer fits.Close()

	// create primary HDU image
	var (
		bitpix = 8
		axes   = []int{3, 4}
		data   = []int8{
			0, 1, 2, 3,
			4, 5, 6, 7,
			8, 9, 0, 1,
		}
	)
	img := fitsio.NewImage(bitpix, axes)
	defer img.Close()

	err = img.Header().Append(
		fitsio.Card{
			"EXTNAME",
			"primary hdu",
			"the primary HDU",
		},
		fitsio.Card{
			"EXTVER",
			2,
			"the primary hdu version",
		},
	)
	if err != nil {
		log.Fatalf("could append cards: %+v", err)
	}

	err = img.Write(data)
	if err != nil {
		log.Fatalf("could not write data to image: %+v", err)
	}

	err = fits.Write(img)
	if err != nil {
		log.Fatalf("could not write image to FITS file: %+v", err)
	}

	err = fits.Close()
	if err != nil {
		log.Fatalf("could not close FITS file: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("could not close file: %+v", err)
	}

	// Output:
}

func Example_createImage3x4I64() {
	f, err := os.Create("testdata/ex-img-bitpix+64.fits")
	if err != nil {
		log.Fatalf("could not create file: %+v", err)
	}
	defer f.Close()

	fits, err := fitsio.Create(f)
	if err != nil {
		log.Fatalf("could not create FITS file: %+v", err)
	}
	defer fits.Close()

	// create primary HDU image
	var (
		bitpix = 64
		axes   = []int{3, 4}
		data   = []int64{
			0, 1, 2, 3,
			4, 5, 6, 7,
			8, 9, 0, 1,
		}
	)
	img := fitsio.NewImage(bitpix, axes)
	defer img.Close()

	err = img.Header().Append(
		fitsio.Card{
			"EXTNAME",
			"primary hdu",
			"the primary HDU",
		},
		fitsio.Card{
			"EXTVER",
			2,
			"the primary hdu version",
		},
	)
	if err != nil {
		log.Fatalf("could append cards: %+v", err)
	}

	err = img.Write(data)
	if err != nil {
		log.Fatalf("could not write data to image: %+v", err)
	}

	err = fits.Write(img)
	if err != nil {
		log.Fatalf("could not write image to FITS file: %+v", err)
	}

	err = fits.Close()
	if err != nil {
		log.Fatalf("could not close FITS file: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("could not close file: %+v", err)
	}

	// Output:
}

func Example_createImage3x4F64() {
	f, err := os.Create("testdata/ex-img-bitpix-64.fits")
	if err != nil {
		log.Fatalf("could not create file: %+v", err)
	}
	defer f.Close()

	fits, err := fitsio.Create(f)
	if err != nil {
		log.Fatalf("could not create FITS file: %+v", err)
	}
	defer fits.Close()

	// create primary HDU image
	var (
		bitpix = -64
		axes   = []int{3, 4}
		data   = []float64{
			0, 1, 2, 3,
			4, 5, 6, 7,
			8, 9, 0, 1,
		}
	)
	img := fitsio.NewImage(bitpix, axes)
	defer img.Close()

	err = img.Header().Append(
		fitsio.Card{
			"EXTNAME",
			"primary hdu",
			"the primary HDU",
		},
		fitsio.Card{
			"EXTVER",
			2,
			"the primary hdu version",
		},
	)
	if err != nil {
		log.Fatalf("could append cards: %+v", err)
	}

	err = img.Write(data)
	if err != nil {
		log.Fatalf("could not write data to image: %+v", err)
	}

	err = fits.Write(img)
	if err != nil {
		log.Fatalf("could not write image to FITS file: %+v", err)
	}

	err = fits.Close()
	if err != nil {
		log.Fatalf("could not close FITS file: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("could not close file: %+v", err)
	}

	// Output:
}
