package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	fits "github.com/astrogo/fitsio"
)

func main() {
	rc := run()
	os.Exit(rc)
}

func run() int {
	var single bool

	flag.Usage = func() {
		const msg = `Usage: go-fitsio-listhead filename[ext]


List the FITS header keywords in a single extension, or, if 
ext is not given, list the keywords in all the extensions. 

Examples:
   
   go-fitsio-listhead file.fits      - list every header in the file 
   go-fitsio-listhead file.fits[0]   - list primary array header 
   go-fitsio-listhead file.fits[2]   - list header of 2nd extension 
   go-fitsio-listhead file.fits+2    - same as above 
   go-fitsio-listhead file.fits[GTI] - list header of GTI extension

Note that it may be necessary to enclose the input file
name in single quote characters on the Unix command line.
`
		fmt.Fprintf(os.Stderr, "%v\n", msg)
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return 1
	}

	fname := flag.Arg(0)
	r, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** %v\n", err)
		return 1
	}
	defer r.Close()

	f, err := fits.Open(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error** %v\n", err)
		return 1
	}
	defer f.Close()

	// list only a single header if a specific extension was given
	if strings.Contains(fname, "[") {
		single = true
	}

	for i := 0; i < len(f.HDUs()); i++ {
		hdu := f.HDU(i)
		hdr := hdu.Header()
		fmt.Printf("Header listing for HDU #%d:\n", i)

		for k := range hdr.Keys() {
			card := hdr.Card(k)
			fmt.Printf(
				"%-8s= %-29s / %s\n",
				card.Name,
				fmt.Sprintf("%v", card.Value),
				card.Comment)
		}
		fmt.Printf("END\n\n")

		// quit if only listing a single header
		if single {
			break
		}
	}
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "**error** %v\n", err)
		return 1
	}

	return 0
}
