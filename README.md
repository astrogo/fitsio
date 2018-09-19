fitsio
======

[![Build Status](https://travis-ci.org/astrogo/fitsio.svg?branch=master)](https://travis-ci.org/astrogo/fitsio)
[![codecov](https://codecov.io/gh/astrogo/fitsio/branch/master/graph/badge.svg)](https://codecov.io/gh/astrogo/fitsio)
[![GoDoc](https://godoc.org/github.com/astrogo/fitsio?status.svg)](https://godoc.org/github.com/astrogo/fitsio)

`fitsio` is a pure-Go package to read and write `FITS` files.

## Installation

```sh
$ go get github.com/astrogo/fitsio
```

## Documentation

http://godoc.org/github.com/astrogo/fitsio

## Contribute

`astrogo/fitsio` is released under `BSD-3`.
Please send a pull request to [astrogo/license](https://github.com/astrogo/license),
adding yourself to the `AUTHORS` and/or `CONTRIBUTORS` file.

## Example

```go
import fits "github.com/astrogo/fitsio"

func dumpFitsTable(fname string) {
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

	// get the second HDU
	table := f.HDU(1).(*fits.Table)
	nrows := table.NumRows()
    rows, err := table.Read(0, nrows)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	for rows.Next() {
        var x, y float64
        var id int64
        err = rows.Scan(&id, &x, &y)
        if err != nil {
            panic(err)
        }
        fmt.Printf(">>> %v %v %v\n", id, x, y)
	}
    err = rows.Err()
    if err != nil { panic(err) }
    
    // using a struct
    xx := struct{
        Id int     `fits:"ID"`
        X  float64 `fits:"x"`
        Y  float64 `fits:"y"`
    }{}
    // using a map
    yy := make(map[string]interface{})
    
    rows, err = table.Read(0, nrows)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
	for rows.Next() {
        err = rows.Scan(&xx)
        if err != nil {
            panic(err)
        }
        fmt.Printf(">>> %v\n", xx)

        err = rows.Scan(&yy)
        if err != nil {
            panic(err)
        }
        fmt.Printf(">>> %v\n", yy)
	}
    err = rows.Err()
    if err != nil { panic(err) }
    
}

```

## TODO

- ``[DONE]`` add support for writing tables from structs
- ``[DONE]`` add support for writing tables from maps
- ``[DONE]`` add support for variable length array
- provide benchmarks _wrt_ ``CFITSIO``
- add support for `TUNITn`
- add support for `TSCALn`
- add suport for `TDIMn`
- add support for (fast, low-level) copy of `FITS` tables

