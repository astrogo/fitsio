//+build ignore

package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"os"

	"github.com/astrogo/fitsio"
	"github.com/astrogo/fitsio/fltimg"
)

var (
	bitpix  = flag.Int("bitpix", 8, "bitpix")
	width   = flag.Int("width", 50, "image width")
	height  = flag.Int("height", 50, "image height")
	nimages = flag.Int("n", 2, "number of images per file")
)

func main() {
	flag.Parse()

	fname := "file.fits"
	w, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	f, err := fitsio.Create(w)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rect := image.Rect(0, 0, *width, *height)

	for i := 0; i < *nimages; i++ {
		img := fitsio.NewImage(*bitpix, []int{*width, *height})
		if img == nil {
			log.Fatalf("error creating new image")
		}
		defer img.Close()

		err = img.Header().Append(fitsio.Card{"IMAGENBR", i, "image number"})
		if err != nil {
			log.Fatalf("error creating card for image #%d: %v\n", i, err)
		}

		switch *bitpix {
		case 8:
			raw := image.NewGray(rect)
			setImage(i, raw, *width, *height)
			err = img.Write(&raw.Pix)
			if err != nil {
				log.Fatalf("error writing pixels: %v\n", err)
			}

		case 16:
			raw := image.NewGray16(rect)
			setImage(i, raw, *width, *height)
			pix := make([]int16, len(raw.Pix)/2)
			for i := 0; i < len(raw.Pix); i += 2 {
				buf := raw.Pix[i : i+2]
				pix[i/2] = int16(uint16(buf[1]) | uint16(buf[0])<<8)
			}
			err := img.Write(&pix)
			if err != nil {
				log.Fatalf("error writing pixels: %v\n", err)
			}

		case 32:
			raw := image.NewRGBA(rect)
			setImage(i, raw, *width, *height)
			pix := make([]int32, len(raw.Pix)/4)
			for i := 0; i < len(raw.Pix); i += 4 {
				buf := raw.Pix[i : i+4]
				pix[i/4] = int32(uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24)
			}
			err := img.Write(&pix)
			if err != nil {
				log.Fatalf("error writing pixels: %v\n", err)
			}

		case 64:
			raw := image.NewRGBA64(rect)
			setImage(i, raw, *width, *height)
			pix := make([]int64, len(raw.Pix)/8)
			for i := 0; i < len(raw.Pix); i += 8 {
				buf := raw.Pix[i : i+8]
				pix[i/8] = int64(uint64(buf[7]) | uint64(buf[6])<<8 | uint64(buf[5])<<16 | uint64(buf[4])<<24 |
					uint64(buf[3])<<32 | uint64(buf[2])<<40 | uint64(buf[1])<<48 | uint64(buf[0])<<56)
			}
			err := img.Write(&pix)
			if err != nil {
				log.Fatalf("error writing pixels: %v\n", err)
			}

		case -32:
			raw := fltimg.NewGray32(rect, make([]byte, 4**width**height))
			setImage(i, raw, *width, *height)
			pix := make([]float32, len(raw.Pix)/4)
			for i := 0; i < len(raw.Pix); i += 4 {
				buf := raw.Pix[i : i+4]
				pix[i/4] = math.Float32frombits(uint32(buf[3]) | uint32(buf[2])<<8 | uint32(buf[1])<<16 | uint32(buf[0])<<24)
			}
			err := img.Write(&pix)
			if err != nil {
				log.Fatalf("error writing pixels: %v\n", err)
			}

		case -64:
			raw := fltimg.NewGray64(rect, make([]byte, 8**width**height))
			setImage(i, raw, *width, *height)
			pix := make([]float64, len(raw.Pix)/8)
			for i := 0; i < len(raw.Pix); i += 8 {
				buf := raw.Pix[i : i+8]
				pix[i/8] = math.Float64frombits(uint64(buf[7]) | uint64(buf[6])<<8 | uint64(buf[5])<<16 | uint64(buf[4])<<24 |
					uint64(buf[3])<<32 | uint64(buf[2])<<40 | uint64(buf[1])<<48 | uint64(buf[0])<<56)
			}
			err := img.Write(&pix)
			if err != nil {
				log.Fatalf("error writing pixels: %v\n", err)
			}

		default:
			log.Fatalf("invalid bitpix value (%d)", *bitpix)
		}

		err = f.Write(img)
		if err != nil {
			log.Fatalf("error writing image #%d: %v\n", i, err)
		}
	}

}

type Circle struct {
	X, Y, R float64
}

func (c *Circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		// outside
		return 0
	} else {
		// inside
		return uint8((1 - math.Pow(d, 5)) * 255)
	}
}

func setImage(i int, m draw.Image, w, h int) {
	hw := float64(w / 2)
	hh := float64(h / 2)
	r1 := math.Min(hw, hh) / 3
	r2 := math.Min(hw, hh) / 2

	θ := 2 * math.Pi / 3
	cr := &Circle{hw - r1*math.Sin(0), hh - r1*math.Cos(0), r2}
	cg := &Circle{hw - r1*math.Sin(θ), hh - r1*math.Cos(θ), r2}
	cb := &Circle{hw - r1*math.Sin(-θ), hh - r1*math.Cos(-θ), r2}

	var circles [3]*Circle
	switch i % 2 {
	case 0:
		circles = [3]*Circle{cr, cg, cb}
	default:
		circles = [3]*Circle{cg, cb, cr}
	}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := color.RGBA{
				circles[0].Brightness(float64(x), float64(y)),
				circles[1].Brightness(float64(x), float64(y)),
				circles[2].Brightness(float64(x), float64(y)),
				255,
			}
			m.Set(x, y, c)
		}
	}
}
