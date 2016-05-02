// Copyright 2016 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"

	"github.com/astrogo/fitsio"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/paint"
)

var (
	infos []fileInfo
	help  = flag.Bool("help", false, "show help")
)

type fileInfo struct {
	Name   string
	Images []image.Image
}

func main() {

	var r *os.File
	var f *fitsio.File
	var err error

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `view-fits - a FITS image viewer.

Usage of view-fits:
$ view-fits [file-names]

Examples:
$ view-fits astrogo/fitsio/testdata/file-img2-bitpix+08.fits
$ view-fits astrogo/fitsio/testdata/file-img2-bitpix+08.fits astrogo/fitsio/testdata/file-img2-bitpix+16.fits astrogo/fitsio/testdata/file-img2-bitpix+64.fits

Navigation:
right/left arrows: switch to next/previous file
up/down arrows: switch to previous/next image in the current file
`)
	}

	flag.Parse()

	if *help || len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	// Parsing input files.
	for _, fname := range flag.Args() {

		finfo := fileInfo{Name: fname}

		// Opening the file.
		if r, err = os.Open(fname); err != nil {
			log.Fatal("Can not open the input file: %s.", err)
		}

		// Opening the FITS file.
		if f, err = fitsio.Open(r); err != nil {
			log.Fatal("Can not open the FITS input file: %s.", err)
		}
		defer f.Close()

		// Getting the file HDUs.
		hdus := f.HDUs()
		for _, hdu := range hdus {
			// Getting the header informations.
			header := hdu.Header()
			axes := header.Axes()

			// Discarding HDU with no axes.
			if len(axes) != 0 {
				// Keeping HDU with IMAGE_HDU type.
				if HDUType := hdu.Type(); HDUType == fitsio.IMAGE_HDU {
					// Filling the images array.
					img := hdu.(fitsio.Image).Image()
					if img != nil {
						finfo.Images = append(finfo.Images, img)
					}
				}
			}
		}

		if len(finfo.Images) > 0 {
			infos = append(infos, finfo)
		}
	}

	if len(infos) == 0 {
		log.Fatal("No image among given FITS files.")
	}

	driver.Main(func(s screen.Screen) {

		// Number of files.
		nbFiles := len(infos)

		// Current displayed file and image in file.
		ifile := 0
		iimg := 0

		// Building the main window.
		w, err := s.NewWindow(&screen.NewWindowOptions{})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()
		// Building the screen buffer.
		b, err := s.NewBuffer(image.Point{})
		if err != nil {
			log.Fatal(err)
		}
		defer b.Release()

		w.Fill(b.Bounds(), color.Black, draw.Src)
		w.Publish()

		for {
			e := w.NextEvent()

			switch e := e.(type) {
			default:
			case key.Event:
				switch e.Code {
				case key.CodeEscape, key.CodeQ:
					return
				case key.CodeRightArrow:
					if e.Direction == key.DirPress {
						if ifile < nbFiles-1 {
							ifile++
						} else {
							ifile = 0
						}
						iimg = 0
						b.Release()
						w.Send(paint.Event{})
					}
				case key.CodeLeftArrow:
					if e.Direction == key.DirPress {
						if ifile == 0 {
							ifile = nbFiles - 1
						} else {
							ifile--
						}
						iimg = 0
						b.Release()
						w.Send(paint.Event{})
					}

				case key.CodeDownArrow:
					if e.Direction == key.DirPress {
						nbImg := len(infos[ifile].Images)
						if iimg < nbImg-1 {
							iimg++
						} else {
							iimg = 0
						}
						b.Release()
						w.Send(paint.Event{})

					}
				case key.CodeUpArrow:
					if e.Direction == key.DirPress {
						nbImg := len(infos[ifile].Images)
						if iimg == 0 {
							iimg = nbImg - 1
						} else {
							iimg--
						}
						b.Release()
						w.Send(paint.Event{})

					}
				}

			case paint.Event:
				img := infos[ifile].Images[iimg]

				b, err = s.NewBuffer(img.Bounds().Size())
				if err != nil {
					log.Fatal(err)
				}
				defer b.Release()

				draw.Draw(b.RGBA(), b.Bounds(), img, image.Point{}, draw.Src)

				w.Upload(image.Point{}, b, img.Bounds())
				w.Publish()
			}

		}

	})
}
