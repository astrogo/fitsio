package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	fits "github.com/astrogo/fitsio"
)

func main() {

	flag.Usage = func() {
		const msg = `Usage: go-fitsio-mergefiles -o outfname file1 file2 [file3 ...]

Merge FITS tables into a single file/table.

`
		fmt.Fprintf(os.Stderr, "%v\n", msg)
		flag.PrintDefaults()
	}

	outfname := flag.String("o", "out.fits", "path to merged FITS file")

	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	_, err := os.Stat(*outfname)
	if err == nil {
		err = os.Remove(*outfname)
		if err != nil {
			panic(err)
		}
	}

	start := time.Now()
	defer func() {
		delta := time.Since(start)
		fmt.Printf("::: timing: %v\n", delta)
	}()

	fmt.Printf("::: creating merged file [%s]...\n", *outfname)
	w, err := os.Create(*outfname)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	out, err := fits.Create(w)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	infiles := make([]string, 0, flag.NArg())
	for i := 0; i < flag.NArg(); i++ {
		fname := flag.Arg(i)
		infiles = append(infiles, fname)
	}

	var table *fits.Table
	fmt.Printf("::: merging [%d] FITS files...\n", len(infiles))
	for i, fname := range infiles {
		r, err := os.Open(fname)
		if err != nil {
			panic(err)
		}
		defer r.Close()
		f, err := fits.Open(r)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		hdu := f.HDU(1).(*fits.Table)
		nrows := hdu.NumRows()
		fmt.Printf("::: reading [%s] -> nrows=%d\n", fname, nrows)
		if i == 0 {
			// get header from first input file
			err = fits.CopyHDU(out, f, 0)
			if err != nil {
				panic(err)
			}

			// get schema from first input file
			cols := hdu.Cols()
			table, err = fits.NewTable(hdu.Name(), cols, hdu.Type())
			if err != nil {
				panic(err)
			}
			defer table.Close()
		}

		err = fits.CopyTable(table, hdu)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("::: merging [%d] FITS files... [done]\n", len(infiles))
	fmt.Printf("::: nrows: %d\n", table.NumRows())
}
