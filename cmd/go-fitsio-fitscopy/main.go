package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	fits "github.com/astrogo/fitsio"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage = func() {
			const msg = `Usage:  go-fitsio-fitscopy inputfile outputfile

Copy an input file to an output file, optionally filtering
the file in the process.  This seemingly simple program can
apply powerful filters which transform the input file as
it is being copied.  Filters may be used to extract a
subimage from a larger image, select rows from a table,
filter a table with a GTI time extension or a SAO region file,
create or delete columns in a table, create an image by
binning (histogramming) 2 table columns, and convert IRAF
format *.imh or raw binary data files into FITS images.
See the CFITSIO User's Guide for a complete description of
the Extended File Name filtering syntax.

Examples:

go-fitsio-fitscopy in.fit out.fit                   (simple file copy)
go-fitsio-fitscopy - -                              (stdin to stdout)
go-fitsio-fitscopy in.fit[11:50,21:60] out.fit      (copy a subimage)
go-fitsio-fitscopy iniraf.imh out.fit               (IRAF image to FITS)
go-fitsio-fitscopy in.dat[i512,512] out.fit         (raw array to FITS)
go-fitsio-fitscopy in.fit[events][pi>35] out.fit    (copy rows with pi>35)
go-fitsio-fitscopy in.fit[events][bin X,Y] out.fit  (bin an image) 
go-fitsio-fitscopy in.fit[events][col x=.9*y] out.fit        (new x column)
go-fitsio-fitscopy in.fit[events][gtifilter()] out.fit       (time filter)
go-fitsio-fitscopy in.fit[2][regfilter("pow.reg")] out.fit (spatial filter)

Note that it may be necessary to enclose the input file name
in single quote characters on the Unix command line.
`
			fmt.Fprintf(os.Stderr, "%v\n", msg)
		}
		flag.Usage()
		os.Exit(1)
	}
	var err error

	ifname := flag.Arg(0)
	ofname := flag.Arg(1)

	// open input file
	var r io.Reader
	if ifname == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(ifname)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		r = f
	}

	in, err := fits.Open(r)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	// create output file
	var w io.WriteCloser
	if ofname == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(ofname)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w = f
	}
	out, err := fits.Create(w)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// copy every HDU until we get an error
	for ihdu := range in.HDUs() {
		err = fits.CopyHDU(out, in, ihdu)
		if err != nil {
			panic(err)
		}
	}

	err = out.Close()
	if err != nil {
		log.Fatalf("could not close output FITS file: %v", err)
	}

	err = w.Close()
	if err != nil {
		log.Fatalf("could not close output file: %v", err)
	}
}
