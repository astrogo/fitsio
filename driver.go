package fitsio

import (
	"fmt"
	"io"
)

var g_drivers map[string]Driver

type Driver interface {
	// Open opens an already existing file for reading.
	Open(fname string) (Conn, error)

	// OpenFile opens an already existing file for reading and/or writing
	OpenFile(fname string, mode Mode) (Conn, error)

	// Create creates a new file for writing.
	Create(fname string) (Conn, error)

	// Name returns the name of the Driver
	Name() string
}

// Conn is a generic connection to a FITS file.
type Conn interface {
	Name() string
	io.Reader
	io.Writer
	io.Closer
}

// Register makes a FITS file driver available.
// If Register is called twice with the same driver or if driver is nil
// it panics.
func Register(driver Driver) {
	if driver == nil {
		panic(fmt.Errorf("fitsio.Register: nil driver"))
	}

	name := driver.Name()
	_, dup := g_drivers[name]
	if dup {
		panic(fmt.Errorf("fitsio.Register: duplicate driver [%s]", name))
	}

	g_drivers[name] = driver
}

func init() {
	g_drivers = make(map[string]Driver, 1)
}
