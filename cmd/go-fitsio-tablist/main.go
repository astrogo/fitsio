package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	fits "github.com/astrogo/fitsio"
)

func main() {
	rc := run()
	os.Exit(rc)
}

func run() int {

	flag.Usage = func() {
		const msg = `Usage: go-fitsio-tablist filename[ext][col filter][row filter]

List the contents of a FITS table

Examples:
  tablist tab.fits[GTI]           - list the GTI extension
  tablist tab.fits[1][#row < 101] - list first 100 rows
  tablist tab.fits[1][col X;Y]    - list X and Y cols only
  tablist tab.fits[1][col -PI]    - list all but the PI col
  tablist tab.fits[1][col -PI][#row < 101]  - combined case

Display formats can be modified with the TDISPn keywords.
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer r.Close()

	f, err := fits.Open(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer f.Close()

	for _, hdu := range f.HDUs() {
		if hdu.Type() == fits.IMAGE_HDU {
			//fmt.Fprintf(os.Stderr, "Error: this program only displays tables, not images\n")
			//return 1
			continue
		}

		table := hdu.(*fits.Table)
		ncols := len(table.Cols())
		nrows := table.NumRows()
		rows, err := table.Read(0, nrows)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		w := os.Stdout
		hdrline := strings.Repeat("=", 80-15)
		maxname := 10
		for _, col := range table.Cols() {
			if len(col.Name) > maxname {
				maxname = len(col.Name)
			}
		}

		data := make([]interface{}, ncols)
		names := make([]string, ncols)
		for i, col := range table.Cols() {
			names[i] = col.Name
			data[i] = reflect.New(col.Type()).Interface()
		}

		rowfmt := fmt.Sprintf("%%-%ds | %%v\n", maxname)
		for irow := 0; rows.Next(); irow++ {
			err = rows.Scan(data...)
			if err != nil {
				fmt.Printf("Error: (row=%v) %v\n", irow, err)
			}
			fmt.Fprintf(w, "== %05d/%05d %s\n", irow, nrows, hdrline)
			for i := 0; i < ncols; i++ {
				rv := reflect.Indirect(reflect.ValueOf(data[i]))
				fmt.Fprintf(w, rowfmt, names[i], rv.Interface())
			}
		}

		err = rows.Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	return 0
}
