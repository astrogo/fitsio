package fitsio

import (
	"fmt"
	"io"
	"os"
)

// Mode defines a FITS file access mode (r/w)
type Mode int

const (
	ReadOnly  Mode = Mode(os.O_RDONLY) // open the file read-only
	WriteOnly      = Mode(os.O_WRONLY) // open the file write-only
	ReadWrite      = Mode(os.O_RDWR)   // open the file read-write
)

// File represents a FITS file.
type File struct {
	dec     Decoder
	enc     Encoder
	name    string
	mode    Mode
	hdus    []HDU
	closers []io.Closer
}

// Open opens a FITS file in read-only mode.
func Open(r io.Reader) (*File, error) {
	var err error

	type namer interface {
		Name() string
	}
	name := ""
	if r, ok := r.(namer); ok {
		name = r.Name()
	}

	f := &File{
		dec:     NewDecoder(r),
		name:    name,
		mode:    ReadOnly,
		hdus:    make([]HDU, 0, 1),
		closers: make([]io.Closer, 0, 1),
	}

	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	if rr, ok := r.(io.Closer); ok {
		f.closers = append(f.closers, rr)
	}

	for {
		var hdu HDU
		hdu, err = f.dec.DecodeHDU()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			err = nil
			break
		}
		f.hdus = append(f.hdus, hdu)
		f.closers = append(f.closers, hdu)
	}

	return f, err
}

// Create creates a new FITS file in write-only mode
func Create(w io.Writer) (*File, error) {
	var err error
	type namer interface {
		Name() string
	}
	name := ""
	if w, ok := w.(namer); ok {
		name = w.Name()
	}

	f := &File{
		enc:     NewEncoder(w),
		name:    name,
		mode:    WriteOnly,
		hdus:    make([]HDU, 0, 1),
		closers: make([]io.Closer, 0, 1),
	}

	if ww, ok := w.(io.Closer); ok {
		f.closers = append(f.closers, ww)
	}

	return f, err
}

// Close releases resources held by a FITS file
func (f *File) Close() error {
	var err error

	for _, v := range f.closers {
		err = v.Close()
		if err != nil {
			return err
		}
	}
	return err
}

// Mode returns the access-mode of this FITS file
func (f *File) Mode() Mode {
	return f.mode
}

// Name returns the name of the FITS file
func (f *File) Name() string {
	return f.name
}

// HDUs returns the list of all Header-Data Unit blocks in the file
func (f *File) HDUs() []HDU {
	return f.hdus
}

// HDU returns the i-th HDU
func (f *File) HDU(i int) HDU {
	return f.hdus[i]
}

// Get returns the HDU with name `name` or nil
func (f *File) Get(name string) HDU {
	i, hdu := f.gethdu(name)
	if i < 0 {
		return nil
	}
	return hdu
}

// Has returns whether the File has a HDU with name `name`.
func (f *File) Has(name string) bool {
	i, _ := f.gethdu(name)
	if i < 0 {
		return false
	}
	return true
}

// get returns the index and HDU of HDU with name `name`.
func (f *File) gethdu(name string) (int, HDU) {
	for i, hdu := range f.hdus {
		if hdu.Name() == name {
			return i, hdu
		}
	}
	return -1, nil
}

// Write writes a HDU to file
func (f *File) Write(hdu HDU) error {
	var err error
	if f.mode != WriteOnly && f.mode != ReadWrite {
		return fmt.Errorf("fitsio: file not open for write")
	}

	if len(f.hdus) == 0 {
		if hdu.Type() != IMAGE_HDU {
			return fmt.Errorf("fitsio: file has no primary header. create one first")
		}

		hdr := hdu.Header()
		if hdr.Get("SIMPLE") == nil {
			err = hdr.prepend(Card{
				Name:    "SIMPLE",
				Value:   true,
				Comment: "primary HDU",
			})
			if err != nil {
				return err
			}
		}
	} else {
		switch hdu.Type() {
		case IMAGE_HDU:
			img := hdu.(Image)
			err = img.freeze()
			if err != nil {
				return err
			}

		case ASCII_TBL, BINARY_TBL:
			tbl := hdu.(*Table)
			err = tbl.freeze()
			if err != nil {
				return err
			}
		}
	}

	err = f.enc.EncodeHDU(hdu)
	if err != nil {
		return err
	}

	err = f.append(hdu)
	if err != nil {
		return err
	}

	return err
}

// append appends an HDU to the list of Header-Data Unit blocks.
func (f *File) append(hdu HDU) error {
	var err error
	if f.mode != WriteOnly && f.mode != ReadWrite {
		return fmt.Errorf("fitsio: file not open for write")
	}

	// mare sure there is only one primary-hdu
	if _, ok := hdu.(*primaryHDU); ok && len(f.hdus) != 0 {
		return fmt.Errorf("fitsio: file has already a Primary HDU")
	}

	f.hdus = append(f.hdus, hdu)
	f.closers = append(f.closers, hdu)
	return err
}
