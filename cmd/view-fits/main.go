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
	"github.com/nfnt/resize"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type fileInfo struct {
	Name   string
	Images []imageInfo
}

type imageInfo struct {
	image.Image
	scale int // image scale in percents (default: 100%)
	orig  image.Point
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
- +:                 increase zoom-level by 20%%
- -:                 decrease zoom-level by 20%%
- ?:                 show help
- q/ESC:             quit

Mouse controls:
- Left button:       pan image
- Wheel-Up:          increase zoom-level by 20%%
- Wheel-Down:        decrease zoom-level by 20%%
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

		var (
			sz size.Event
			//bkg = color.Black
			bkg = color.RGBA{0xe0, 0xe0, 0xe0, 0xff} // Material Design "Grey 300"

			mbl image.Rectangle // mouse button-left position

			repaint = true
			panning = false
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

			case mouse.Event:
				ix := int(e.X)
				iy := int(e.Y)
				switch e.Button {
				case mouse.ButtonLeft:
					switch e.Direction {
					case mouse.DirPress:
						panning = true
						mbl = image.Rect(ix, iy, ix, iy)

					case mouse.DirRelease:
						panning = false
						mbl.Max = image.Point{ix, iy}

						switch {
						case e.Modifiers&key.ModShift != 0:
							// zoom-in
						default:
							// pan
							repaint = true
							img := &infos[cur.file].Images[cur.img]
							dx := mbl.Dx()
							dy := mbl.Dy()
							img.orig = originTrans(img.orig.Sub(image.Point{dx, dy}), sz.Bounds(), img)
						}
					}

				case mouse.ButtonRight:
					// no-op

				case mouse.ButtonWheelDown:
					if e.Direction == mouse.DirStep {
						ctrlZoomOut(&infos[cur.file].Images[cur.img], &repaint)
					}
				case mouse.ButtonWheelUp:
					if e.Direction == mouse.DirStep {
						ctrlZoomIn(&infos[cur.file].Images[cur.img], &repaint)
					}

				}

				if panning {
					repaint = true
					img := &infos[cur.file].Images[cur.img]
					mbl.Max = image.Point{ix, iy}
					dx := mbl.Dx()
					dy := mbl.Dy()
					img.orig = originTrans(img.orig.Sub(image.Point{dx, dy}), sz.Bounds(), img)
					mbl.Min = mbl.Max
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

				case key.CodeKeypadPlusSign, key.CodeEqualSign:
					if e.Direction == key.DirPress {
						ctrlZoomIn(&infos[cur.file].Images[cur.img], &repaint)
					}

				case key.CodeHyphenMinus:
					if e.Direction == key.DirPress {
						ctrlZoomOut(&infos[cur.file].Images[cur.img], &repaint)
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

			case size.Event:
				sz = e

			case paint.Event:
				if !repaint {
					continue
				}
				repaint = false
				img := infos[cur.file].Images[cur.img]

				release(b)
				curImg := img.get()
				b, err = s.NewBuffer(curImg.Bounds().Size())
				if err != nil {
					log.Fatal(err)
				}
				defer release(b)

				draw.Draw(b.RGBA(), b.Bounds(), curImg, img.orig, draw.Src)

				w.Fill(sz.Bounds(), bkg, draw.Src)
				w.Upload(image.Point{}, b, curImg.Bounds())
				w.Publish()
			}

			if repaint {
				w.Send(paint.Event{})
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
						finfo.Images = append(finfo.Images, imageInfo{
							Image: img,
							scale: 100,
							orig:  image.Point{},
						})
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

func ctrlZoomOut(img *imageInfo, repaint *bool) {
	*repaint = true
	if img.scale <= 20 {
		*repaint = false
		return
	}

	img.scale -= 20

	if img.scale <= 0 {
		img.scale = 20
	}
}

func ctrlZoomIn(img *imageInfo, repaint *bool) {
	*repaint = true
	img.scale += 20
}

func (img *imageInfo) get() image.Image {
	if img.scale == 100 {
		return img.Image
	}
	width := uint(float64(img.scale) / 100.0 * float64(img.Bounds().Dx()))
	height := uint(float64(img.scale) / 100.0 * float64(img.Bounds().Dy()))
	return resize.Resize(width, height, img.Image, resize.MitchellNetravali)
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// originTrans translates the origin with respect to the current image and the
// current canvas size. This makes sure we never incorrect position the image.
// (i.e., panning never goes too far, and whenever the canvas is bigger than
// the image, the origin is *always* (0, 0).
func originTrans(pt image.Point, win image.Rectangle, img *imageInfo) image.Point {
	// If there's no valid image, then always return (0, 0).
	if img == nil {
		return image.Point{0, 0}
	}

	// Quick aliases.
	ww := win.Dx()
	wh := win.Dy()
	dw := img.Bounds().Dx() - ww
	dh := img.Bounds().Dy() - wh

	// Set the allowable range of the origin point of the image.
	// i.e., never less than (0, 0) and never greater than the width/height
	// of the image that isn't viewable at any given point (which is determined
	// by the canvas size).
	pt.X = min(img.Bounds().Min.X+dw, max(pt.X, 0))
	pt.Y = min(img.Bounds().Min.Y+dh, max(pt.Y, 0))

	// Validate origin point. If the width/height of an image is smaller than
	// the canvas width/height, then the image origin cannot change in x/y
	// direction.
	if img.Bounds().Dx() < ww {
		pt.X = 0
	}
	if img.Bounds().Dy() < wh {
		pt.Y = 0
	}

	return pt
}
