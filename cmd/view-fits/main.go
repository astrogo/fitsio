// Copyright 2016 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// view-fits is a simple program to display images in a FITS file.
//
// Usage of view-fits:
// $ view-fits [file1 [file2 [...]]]
//
// Examples:
//  $ view-fits astrogo/fitsio/testdata/file-img2-bitpix+08.fits
//  $ view-fits astrogo/fitsio/testdata/file-img2-bitpix*.fits
//  $ view-fits http://data.astropy.org/tutorials/FITS-images/HorseHead.fits
//  $ view-fits file:///some/file.fits
//
// Controls:
//  - left/right arrows: switch to previous/next file
//  - up/down arrows:    switch to previous/next image in the current file
//  - r:                 reload/redisplay current image
//  - z:                 resize window to fit current image
//  - p:                 print current image to 'output.png'
//  - ?:                 show help
//  - q/ESC:             quit
//
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/astrogo/fitsio"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type fileInfo struct {
	Name   string
	Images []image.Image
}

func main() {

	help := flag.Bool("help", false, "show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `view-fits - a FITS image viewer.

Usage of view-fits:
$ view-fits [file1 [file2 [...]]]

Examples:
$ view-fits astrogo/fitsio/testdata/file-img2-bitpix+08.fits
$ view-fits astrogo/fitsio/testdata/file-img2-bitpix*.fits
$ view-fits http://data.astropy.org/tutorials/FITS-images/HorseHead.fits
$ view-fits file:///some/file.fits

Controls:
- left/right arrows: switch to previous/next file
- up/down arrows:    switch to previous/next image in the current file
- r:                 reload/redisplay current image
- z:                 resize window to fit current image
- p:                 print current image to 'output.png'
- ?:                 show help
- q/ESC:             quit
`)
	}

	flag.Parse()

	if *help || len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	log.SetFlags(0)
	log.SetPrefix("[view-fits] ")

	infos := processFiles()
	if len(infos) == 0 {
		log.Fatal("No image among given FITS files.")
	}

	type cursor struct {
		file int
		img  int
	}

	driver.Main(func(s screen.Screen) {

		// Number of files.
		nbFiles := len(infos)

		// Current displayed file and image in file.
		cur := cursor{file: 0, img: 0}

		// Building the main window.
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Width:  500,
			Height: 500,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		// Building the screen buffer.
		b, err := s.NewBuffer(image.Point{500, 500})
		if err != nil {
			log.Fatal(err)
		}
		defer release(b)

		w.Fill(b.Bounds(), color.Black, draw.Src)
		w.Publish()

		repaint := true
		var (
			sz size.Event
			//bkg = color.Black
			bkg = color.RGBA{0xe0, 0xe0, 0xe0, 0xff} // Material Design "Grey 300"
		)

		for {
			switch e := w.NextEvent().(type) {
			default:
				// ignore

			case lifecycle.Event:
				switch {
				case e.From == lifecycle.StageVisible && e.To == lifecycle.StageFocused:
					repaint = true
				default:
					repaint = false
				}
				if repaint {
					w.Send(paint.Event{})
				}

			case key.Event:
				switch e.Code {
				case key.CodeEscape, key.CodeQ:
					return

				case key.CodeSlash:
					if e.Direction == key.DirPress && e.Modifiers&key.ModShift != 0 {
						flag.Usage()
						continue
					}

				case key.CodeRightArrow:
					if e.Direction == key.DirPress {
						repaint = true
						if cur.file < nbFiles-1 {
							cur.file++
						} else {
							cur.file = 0
						}
						cur.img = 0
						log.Printf("file:   %v\n", infos[cur.file].Name)
						log.Printf("images: %d\n", len(infos[cur.file].Images))
					}

				case key.CodeLeftArrow:
					if e.Direction == key.DirPress {
						repaint = true
						if cur.file == 0 {
							cur.file = nbFiles - 1
						} else {
							cur.file--
						}
						cur.img = 0
						log.Printf("file:   %v\n", infos[cur.file].Name)
						log.Printf("images: %d\n", len(infos[cur.file].Images))
					}

				case key.CodeDownArrow:
					if e.Direction == key.DirPress {
						repaint = true
						nbImg := len(infos[cur.file].Images)
						if cur.img < nbImg-1 {
							cur.img++
						} else {
							cur.img = 0
						}
					}

				case key.CodeUpArrow:
					if e.Direction == key.DirPress {
						repaint = true
						nbImg := len(infos[cur.file].Images)
						if cur.img == 0 {
							cur.img = nbImg - 1
						} else {
							cur.img--
						}
					}

				case key.CodeR:
					if e.Direction == key.DirPress {
						repaint = true
					}

				case key.CodeZ:
					if e.Direction == key.DirPress {
						// resize to current image
						// TODO(sbinet)
						repaint = true
					}

				case key.CodeP:
					if e.Direction != key.DirPress {
						continue
					}
					out, err := os.Create("output.png")
					if err != nil {
						log.Fatalf("error printing image: %v\n", err)
					}
					defer out.Close()
					err = png.Encode(out, infos[cur.file].Images[cur.img])
					if err != nil {
						log.Fatalf("error printing image: %v\n", err)
					}
					err = out.Close()
					if err != nil {
						log.Fatalf("error printing image: %v\n", err)
					}
					log.Printf("printed current image to [%s]\n", out.Name())
				}

				if repaint {
					w.Send(paint.Event{})
				}

			case size.Event:
				sz = e

			case paint.Event:
				if !repaint {
					continue
				}
				repaint = false
				img := infos[cur.file].Images[cur.img]

				release(b)
				b, err = s.NewBuffer(img.Bounds().Size())
				if err != nil {
					log.Fatal(err)
				}
				defer release(b)

				draw.Draw(b.RGBA(), b.Bounds(), img, image.Point{}, draw.Src)

				w.Fill(sz.Bounds(), bkg, draw.Src)
				w.Upload(image.Point{}, b, img.Bounds())
				w.Publish()
			}

		}

	})
}

func processFiles() []fileInfo {
	infos := make([]fileInfo, 0, len(flag.Args()))
	// Parsing input files.
	for _, fname := range flag.Args() {

		finfo := fileInfo{Name: fname}

		r, err := openStream(fname)
		if err != nil {
			log.Fatalf("Can not open the input file: %v", err)
		}
		defer r.Close()

		// Opening the FITS file.
		f, err := fitsio.Open(r)
		if err != nil {
			log.Fatalf("Can not open the FITS input file: %v", err)
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
				if hdu, ok := hdu.(fitsio.Image); ok {
					img := hdu.Image()
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

	return infos
}

func openStream(name string) (io.ReadCloser, error) {
	switch {
	case strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://"):
		resp, err := http.Get(name)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		f, err := ioutil.TempFile("", "view-fits-")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			f.Close()
			return nil, err
		}

		// make sure we have at least a full FITS block
		f.Seek(0, 2880)
		f.Seek(0, 0)

		return f, nil

	case strings.HasPrefix(name, "file://"):
		name = name[len("file://"):]
		return os.Open(name)
	default:
		return os.Open(name)
	}
}

type releaser interface {
	Release()
}

func release(r releaser) {
	if r != nil {
		r.Release()
	}
}
