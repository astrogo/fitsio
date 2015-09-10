// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestOpenFile(t *testing.T) {
	const fname = "testdata/file001.fits"
	r, err := os.Open(fname)
	if err != nil {
		t.Fatalf("could not open FITS file [%s]: %v", fname, err)
	}
	defer r.Close()

	f, err := Open(r)
	if err != nil {
		t.Fatalf("could not open FITS file [%s]: %v", fname, err)
	}
	defer f.Close()

	fmode := f.Mode()
	if fmode != ReadOnly {
		t.Fatalf("expected file-mode [%v]. got [%v]", ReadOnly, fmode)
	}

	name := f.Name()
	if name != fname {
		t.Fatalf("expected file-name [%v]. got [%v]", fname, name)
	}

	// furl, err := f.UrlType()
	// if err != nil {
	// 	t.Fatalf("error url: %v", err)
	// }
	// if furl != "file://" {
	// 	t.Fatalf("expected file-url [%v]. got [%v]", "file://", furl)
	// }

}

func TestAsciiTbl(t *testing.T) {
	const fname = "testdata/file001.fits"
	r, err := os.Open(fname)
	if err != nil {
		t.Fatalf("could not open FITS file [%s]: %v", fname, err)
	}
	defer r.Close()

	f, err := Open(r)
	if err != nil {
		t.Fatalf("could not open FITS file [%s]: %v", fname, err)
	}
	defer f.Close()

	nhdus := len(f.HDUs())
	if nhdus != 2 {
		t.Fatalf("expected #hdus [%v]. got [%v]", 2, nhdus)
	}

	hdu := f.HDU(0)
	hdutype := hdu.Type()
	if err != nil {
		t.Fatalf("error hdu-type: %v", err)
	}
	if hdutype != IMAGE_HDU {
		t.Fatalf("expected hdu type [%v]. got [%v]", IMAGE_HDU, hdutype)
	}

	txt := hdu.Header().Text()
	txt_ref, err := ioutil.ReadFile("testdata/ref.file001.hdu")
	if err != nil {
		t.Fatalf("error reading ref file: %v", err)
	}
	if string(txt_ref) != txt {
		t.Fatalf("expected hdu-text:\n%q\ngot:\n%q", string(txt_ref), txt)
	}

	// move to next header
	hdu = f.HDU(1)
	hdutype = hdu.Type()
	if err != nil {
		t.Fatalf("error hdu-type: %v", err)
	}
	if hdutype != ASCII_TBL {
		t.Fatalf("expected hdu type [%v]. got [%v]", ASCII_TBL, hdutype)
	}

	// test hdu by name
	hduok := f.Has("TABLE")
	if hduok {
		t.Fatalf("expected error")
	}
}

func TestBinTable(t *testing.T) {
	const fname = "testdata/swp06542llg.fits"
	r, err := os.Open(fname)
	if err != nil {
		t.Fatalf("could not open FITS file [%s]: %v", fname, err)
	}
	defer r.Close()
	f, err := Open(r)
	if err != nil {
		t.Fatalf("could not open FITS file [%s]: %v", fname, err)
	}
	defer f.Close()

	nhdus := len(f.HDUs())
	if nhdus != 2 {
		t.Fatalf("expected #hdus [%v]. got [%v]", 2, nhdus)
	}

	hdu := f.HDU(0)
	hdutype := hdu.Type()
	if err != nil {
		t.Fatalf("error hdu-type: %v", err)
	}
	if hdutype != IMAGE_HDU {
		t.Fatalf("expected hdu type [%v]. got [%v]", IMAGE_HDU, hdutype)
	}

	txt := hdu.Header().Text()
	txt_ref, err := ioutil.ReadFile("testdata/ref.swp06542llg.hdu")
	if err != nil {
		t.Fatalf("error reading ref file: %v", err)
	}
	if string(txt_ref) != txt {
		t.Fatalf("expected hdu-write:\n%q\ngot:\n%q", string(txt_ref), txt)
	}

	// move to next header
	hdu = f.HDU(1)
	hdutype = hdu.Type()
	if err != nil {
		t.Fatalf("error hdu-type: %v", err)
	}
	if hdutype != BINARY_TBL {
		t.Fatalf("expected hdu type [%v]. got [%v]", BINARY_TBL, hdutype)
	}

	// get hdu by name
	hduok := f.Has("IUE MELO")
	if !hduok {
		t.Fatalf("expected to have a HDU by name 'IUE MELO'")
	}

	hdu = f.Get("IUE MELO")
	if hdu.Type() != BINARY_TBL {
		t.Fatalf("expected hdu 'IUE MELO' to be a bintable. got [%v]", hdu.Type())
	}

}

func TestOpen(t *testing.T) {
	for _, table := range g_tables {
		r, err := os.Open(table.fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", table.fname, err)
		}
		defer r.Close()
		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", table.fname, err)
		}
		defer f.Close()

		fmode := f.Mode()
		if fmode != ReadOnly {
			t.Fatalf("expected file-mode [%v]. got [%v]", ReadOnly, fmode)
		}

		name := f.Name()
		if name != table.fname {
			t.Fatalf("expected file-name [%v]. got [%v]", table.fname, name)
		}

		if len(f.HDUs()) != len(table.hdus) {
			t.Fatalf("#hdus. expected %v. got %v", len(table.hdus), len(f.HDUs()))
		}
	}
}

func TestCreateFile(t *testing.T) {
	curdir, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Chdir(curdir)

	workdir, err := ioutil.TempDir("", "go-fitsio-test-")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.RemoveAll(workdir)

	err = os.Chdir(workdir)
	if err != nil {
		t.Fatalf(err.Error())
	}

	fname := "new.fits"
	w, err := os.Create(fname)
	if err != nil {
		t.Fatalf("error creating new file [%v]: %v", fname, err)
	}
	defer w.Close()
	f, err := Create(w)
	if err != nil {
		t.Fatalf("error creating new file [%v]: %v", fname, err)
	}
	defer f.Close()

	fmode := f.Mode()
	if err != nil {
		t.Fatalf("error mode: %v", err)
	}
	if fmode != WriteOnly {
		t.Fatalf("expected file-mode [%v]. got [%v]", WriteOnly, fmode)
	}

	name := f.Name()
	if err != nil {
		t.Fatalf("error name: %v", err)
	}
	if name != fname {
		t.Fatalf("expected file-name [%v]. got [%v]", fname, name)
	}

	if len(f.HDUs()) != 0 {
		t.Fatalf("#hdus. expected %v. got %v", 0, len(f.HDUs()))
	}

}

// EOF
