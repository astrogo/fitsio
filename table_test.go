// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestTable(t *testing.T) {
	for _, table := range g_tables {
		fname := table.fname
		r, err := os.Open(fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer r.Close()
		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer f.Close()
		for i := range f.HDUs() {
			hdu, ok := f.HDU(i).(*Table)
			if !ok {
				continue
			}
			for irow := int64(0); irow < hdu.NumRows(); irow++ {
				rows, err := hdu.Read(irow, irow+1)
				if err != nil {
					t.Fatalf(
						"error reading row [%v] (fname=%v, table=%v): %v",
						irow, fname, hdu.Name(), err,
					)
				}
				rows.Close()
			}
		}
	}
}

func TestTableNext(t *testing.T) {
	for _, table := range g_tables {
		fname := table.fname
		r, err := os.Open(fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer r.Close()
		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer f.Close()

		for i := range f.HDUs() {
			hdu, ok := f.HDU(i).(*Table)
			if !ok {
				continue
			}

			nrows := hdu.NumRows()
			// iter over all rows
			rows, err := hdu.Read(0, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count := int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}

			// iter over no row
			rows, err = hdu.Read(0, 0)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != 0 {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", 0, count)
			}

			// iter over 1 row
			rows, err = hdu.Read(0, 1)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != 1 {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", 1, count)
			}

			// iter over all rows
			rows, err = hdu.Read(0, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}

			// iter over all rows +1
			rows, err = hdu.Read(0, nrows+1)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}

			// iter over all rows -1
			rows, err = hdu.Read(1, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows-1 {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows-1, count)
			}

			// iter over [1,1+maxrows -1)
			rows, err = hdu.Read(1, nrows-1)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			exp := nrows - 2
			if exp <= 0 {
				exp = 0
			}
			if count != exp {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", exp, count)
			}

			// iter over last row
			rows, err = hdu.Read(nrows-1, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count = int64(0)
			for rows.Next() {
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != 1 {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", 1, count)
			}

		}
	}
}

func TestTableErrScan(t *testing.T) {
	for _, table := range g_tables {
		fname := table.fname
		// fmt.Printf("--- [%s] ---\n", fname)
		r, err := os.Open(fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer r.Close()
		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer f.Close()

		for i := range f.HDUs() {
			hdu, ok := f.HDU(i).(*Table)
			if !ok {
				continue
			}
			// fmt.Printf("### [%s][hdu-%d]...\n", hdu.Name(), i)
			ncols := len(hdu.Cols())
			nrows := hdu.NumRows()
			rows, err := hdu.Read(0, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			defer rows.Close()

			count := int64(0)
			data := make([]interface{}, ncols)
			for i, col := range hdu.Cols() {
				// fmt.Printf(">>> col[%d] %q\n", i, col.Name)
				data[i] = reflect.New(col.Type()).Interface()
			}

			for rows.Next() {
				count++
				err = rows.Scan(data...)
				if err != nil {
					t.Fatalf("rows.Scan: error: %v", err)
				}

				dummy := 0
				err = rows.Scan(&dummy) // none of the tables has only 1 field
				if err == nil {
					t.Fatalf("rows.Scan: expected a failure")
				}
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}
		}
	}
}

func TestTableScan(t *testing.T) {
	for _, table := range g_tables {
		fname := table.fname
		// fmt.Printf(">>> fname=%q\n", fname)
		r, err := os.Open(fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer r.Close()
		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer f.Close()

		for i := range f.HDUs() {
			hdu, ok := f.HDU(i).(*Table)
			if !ok {
				continue
			}
			nrows := hdu.NumRows()
			rows, err := hdu.Read(0, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count := int64(0)
			for rows.Next() {
				ref := make([]interface{}, len(table.tuple[i][count]))
				data := make([]interface{}, len(ref))
				for ii, vv := range table.tuple[i][count] {
					rt := reflect.TypeOf(vv)
					rv := reflect.New(rt)
					xx := rv.Interface()
					data[ii] = xx
					ref[ii] = vv
				}
				err = rows.Scan(data...)
				if err != nil {
					t.Fatalf("rows.Scan: %v", err)
				}
				// check data just read in is ok
				for ii, vv := range data {
					rv := reflect.ValueOf(vv).Elem().Interface()
					if !reflect.DeepEqual(rv, ref[ii]) {
						t.Fatalf("rows.Scan(%s):\nexpected=%v\ngot=%v", fname, ref[ii], rv)
					}
				}
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}
		}
	}
}

func TestTableScanMap(t *testing.T) {
	for _, table := range g_tables {
		fname := table.fname
		r, err := os.Open(fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer r.Close()
		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer f.Close()

		for i := range f.HDUs() {
			hdu, ok := f.HDU(i).(*Table)
			if !ok {
				continue
			}
			refmap := table.maps[i]
			if len(refmap) <= 0 {
				continue
			}
			nrows := hdu.NumRows()
			rows, err := hdu.Read(0, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count := int64(0)
			for rows.Next() {
				ref := make(map[string]interface{}, len(refmap))
				data := make(map[string]interface{}, len(refmap))
				for kk, vv := range table.maps[i] {
					rt := reflect.TypeOf(vv)
					rv := reflect.New(rt)
					xx := rv.Interface()
					data[kk] = xx
					ii := hdu.Index(kk)
					if ii < 0 {
						for _, col := range hdu.cols {
							fmt.Printf(">>> %q\n", col.Name)
						}
						t.Fatalf("could not find index of [%v]\n%v", kk, hdu.colidx)
					}
					ref[kk] = table.tuple[i][count][ii]
				}
				err = rows.Scan(&data)
				if err != nil {
					t.Fatalf("rows.Scan: %v", err)
				}
				// check data just read in is ok
				datakeys := make([]string, 0, len(data))
				for k := range data {
					datakeys = append(datakeys, k)
				}
				refkeys := make([]string, 0, len(ref))
				for k := range ref {
					refkeys = append(refkeys, k)
				}
				sort.Strings(datakeys)
				sort.Strings(refkeys)
				if !reflect.DeepEqual(refkeys, datakeys) {
					t.Fatalf("rows.Scan:\nexpected=%v\ngot=%v", refkeys, datakeys)
				}
				for _, k := range refkeys {
					refval := ref[k]
					//dataval := reflect.ValueOf(data[k]).Elem().Interface()
					dataval := data[k]
					if !reflect.DeepEqual(dataval, refval) {
						t.Fatalf("rows.Scan (key=%s):\nexpected=%v\ngot=%v", k, refval, dataval)
					}
				}
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}
		}
	}
}

func TestTableScanStruct(t *testing.T) {
	for _, table := range g_tables {
		fname := table.fname
		r, err := os.Open(fname)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer r.Close()

		f, err := Open(r)
		if err != nil {
			t.Fatalf("error opening file [%v]: %v", fname, err)
		}
		defer f.Close()

		for i := range f.HDUs() {
			hdu, ok := f.HDU(i).(*Table)
			if !ok {
				continue
			}
			reftypes := table.types[i]
			if reftypes == nil {
				continue
			}
			reftype := reflect.TypeOf(reftypes)
			nrows := hdu.NumRows()
			rows, err := hdu.Read(0, nrows)
			if err != nil {
				t.Fatalf("table.Read: %v", err)
			}
			count := int64(0)
			for rows.Next() {
				ref := reflect.New(reftype)
				data := reflect.New(reftype)
				for ii := 0; ii < reftype.NumField(); ii++ {
					ft := reftype.Field(ii)
					kk := ft.Tag.Get("fits")
					if kk == "" {
						kk = ft.Name
					}
					idx := hdu.Index(kk)
					if idx < 0 {
						t.Fatalf("could not find index of [%v.%v]", reftype.Name(), ft.Name)
					}
					vv := reflect.ValueOf(table.tuple[i][count][idx])
					reflect.Indirect(ref).Field(ii).Set(vv)
				}

				err = rows.Scan(data.Interface())
				if err != nil {
					t.Fatalf("rows.Scan: %v", err)
				}
				// check data just read in is ok
				if !reflect.DeepEqual(data.Interface(), ref.Interface()) {
					t.Fatalf("rows.Scan:\nexpected=%v\ngot=%v", ref, data)
				}
				count++
			}
			err = rows.Err()
			if err != nil {
				t.Fatalf("rows.Err: %v", err)
			}
			if count != nrows {
				t.Fatalf("rows.Next: expected [%d] rows. got %d.", nrows, count)
			}
		}
	}
}

func TestTableBuiltinsRW(t *testing.T) {

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

	for ii, table := range []struct {
		name  string
		cols  []Column
		htype HDUType
		table interface{}
	}{
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int8s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []int8{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int16s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []int16{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int32s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []int32{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int64s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []int64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "ints",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []int{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint8s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []uint8{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint16s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []uint16{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint32s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []uint32{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint64s",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []uint64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uints",
					Format: "I4",
				},
			},
			htype: ASCII_TBL,
			table: []uint{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float32s",
					Format: "E26.17",
				},
			},
			htype: ASCII_TBL,
			table: []float32{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float64s",
					Format: "E26.17",
				},
			},
			htype: ASCII_TBL,
			table: []float64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "strings",
					Format: "A80",
				},
			},
			htype: ASCII_TBL,
			table: []string{
				"10", "11", "12", "13",
				"14", "15", "16", "17",
				"18", "19", "10", "11",
			},
		},
		// binary table
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "bools",
					Format: "L",
				},
			},
			htype: BINARY_TBL,
			table: []bool{
				true, true, true, true,
				false, false, false, false,
				true, false, true, false,
				false, true, false, true,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int8s",
					Format: "B",
				},
			},
			htype: BINARY_TBL,
			table: []int8{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int16s",
					Format: "I",
				},
			},
			htype: BINARY_TBL,
			table: []int16{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int32s",
					Format: "J",
				},
			},
			htype: BINARY_TBL,
			table: []int32{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int64s",
					Format: "K",
				},
			},
			htype: BINARY_TBL,
			table: []int64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "ints",
					Format: "K",
				},
			},
			htype: BINARY_TBL,
			table: []int{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint8s",
					Format: "B",
				},
			},
			htype: BINARY_TBL,
			table: []uint8{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint16s",
					Format: "I",
				},
			},
			htype: BINARY_TBL,
			table: []uint16{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint32s",
					Format: "J",
				},
			},
			htype: BINARY_TBL,
			table: []uint32{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint64s",
					Format: "K",
				},
			},
			htype: BINARY_TBL,
			table: []uint64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uints",
					Format: "K",
				},
			},
			htype: BINARY_TBL,
			table: []uint{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float32s",
					Format: "E",
				},
			},
			htype: BINARY_TBL,
			table: []float32{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float64s",
					Format: "D",
				},
			},
			htype: BINARY_TBL,
			table: []float64{
				10, 11, 12, 13,
				14, 15, 16, 17,
				18, 19, 10, 11,
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "cplx64s",
					Format: "C",
				},
			},
			htype: BINARY_TBL,
			table: []complex64{
				complex(float32(10), float32(10)), complex(float32(11), float32(11)),
				complex(float32(12), float32(12)), complex(float32(13), float32(13)),
				complex(float32(14), float32(14)), complex(float32(15), float32(15)),
				complex(float32(16), float32(16)), complex(float32(17), float32(17)),
				complex(float32(18), float32(18)), complex(float32(19), float32(19)),
				complex(float32(10), float32(10)), complex(float32(11), float32(11)),
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "cplx128s",
					Format: "M",
				},
			},
			htype: BINARY_TBL,
			table: []complex128{
				complex(10, 10), complex(11, 11), complex(12, 12), complex(13, 13),
				complex(14, 14), complex(15, 15), complex(16, 16), complex(17, 17),
				complex(18, 18), complex(19, 19), complex(10, 10), complex(11, 11),
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "strings",
					Format: "10A",
				},
			},
			htype: BINARY_TBL,
			table: []string{
				"10", "11", "12", "13",
				"14", "15", "16", "17",
				"18", "19", "10", "11",
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float64s",
					Format: "2D",
				},
			},
			htype: BINARY_TBL,
			table: [][2]float64{
				{10, 11},
				{12, 13},
				{14, 15},
				{16, 17},
				{18, 19},
				{10, 11},
			},
		},
	} {
		fname := fmt.Sprintf("%03d_%s", ii, table.name)
		for _, fct := range []func(){
			// create
			func() {
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

				phdu, err := NewPrimaryHDU(nil)
				if err != nil {
					t.Fatalf("error creating new primary hdu [%v]: %v", fname, err)
				}
				err = f.Write(phdu)
				if err != nil {
					t.Fatalf("error writing primary hdu [%v]: %v", fname, err)
				}

				tbl, err := NewTable("test", table.cols, table.htype)
				if err != nil {
					t.Fatalf("error creating new table: %v (%v)", err, table.cols[0].Name)
				}
				defer tbl.Close()

				rslice := reflect.ValueOf(table.table)
				for i := 0; i < rslice.Len(); i++ {
					data := rslice.Index(i).Addr()
					err = tbl.Write(data.Interface())
					if err != nil {
						t.Fatalf("error writing row [%v]: %v (data=%v %T)", i, err, data.Interface(), data.Interface())
					}
				}

				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v] (%v)", rslice.Len(), nrows, table.cols[0].Name)
				}

				err = f.Write(tbl)
				if err != nil {
					t.Fatalf("error writing table: %v", err)
				}
			},
			// read
			func() {
				r, err := os.Open(fname)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer r.Close()
				f, err := Open(r)
				if err != nil {
					b, _ := ioutil.ReadFile(fname)
					t.Fatalf("error opening file [%v]: %v\n%q", fname, err, string(b))
				}
				defer f.Close()

				hdu := f.HDU(1)
				tbl := hdu.(*Table)
				if tbl.Name() != "test" {
					t.Fatalf("expected table name==%q. got %q", "test", tbl.Name())
				}

				rslice := reflect.ValueOf(table.table)
				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v]", rslice.Len(), nrows)
				}

				rows, err := tbl.Read(0, nrows)
				if err != nil {
					t.Fatalf("table.Read: %v", err)
				}
				count := int64(0)
				for rows.Next() {
					ref := rslice.Index(int(count)).Interface()
					rt := reflect.TypeOf(ref)
					ptr := reflect.New(rt)
					err = rows.Scan(ptr.Interface())
					if err != nil {
						t.Fatalf("rows.Scan: %v", err)
					}
					data := ptr.Elem().Interface()
					// check data just read in is ok
					if !reflect.DeepEqual(data, ref) {
						t.Fatalf("rows.Scan[row=%[4]d]: %[3]s\nexp=%[1]v (%[1]T)\ngot=%[2]v (%[2]T)", ref, data, table.cols[0].Name, count)
					}
					count++
				}
				if count != nrows {
					t.Fatalf("expected [%v] rows. got [%v]", nrows, count)
				}
			},
		} {
			fct()
		}
	}
}

func TestTableSliceRW(t *testing.T) {

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

	for ii, table := range []struct {
		name  string
		cols  []Column
		htype HDUType
		table interface{}
	}{
		// binary table
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "bools",
					Format: "QL",
				},
			},
			htype: BINARY_TBL,
			table: [][]bool{
				{true, true, true, true},
				{false, false, false, false},
				{true, false, true, false},
				{false, true, false, true},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int8s",
					Format: "QB",
				},
			},
			htype: BINARY_TBL,
			table: [][]int8{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int16s",
					Format: "QI",
				},
			},
			htype: BINARY_TBL,
			table: [][]int16{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int32s",
					Format: "QJ",
				},
			},
			htype: BINARY_TBL,
			table: [][]int32{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int64s",
					Format: "QK",
				},
			},
			htype: BINARY_TBL,
			table: [][]int64{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "ints",
					Format: "QK",
				},
			},
			htype: BINARY_TBL,
			table: [][]int{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint8s",
					Format: "QB",
				},
			},
			htype: BINARY_TBL,
			table: [][]uint8{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint16s",
					Format: "QI",
				},
			},
			htype: BINARY_TBL,
			table: [][]uint16{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint32s",
					Format: "QJ",
				},
			},
			htype: BINARY_TBL,
			table: [][]uint32{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint64s",
					Format: "QK",
				},
			},
			htype: BINARY_TBL,
			table: [][]uint64{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uints",
					Format: "QK",
				},
			},
			htype: BINARY_TBL,
			table: [][]uint{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float32s",
					Format: "QE",
				},
			},
			htype: BINARY_TBL,
			table: [][]float32{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float64s",
					Format: "QD",
				},
			},
			htype: BINARY_TBL,
			table: [][]float64{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "cplx64s",
					Format: "QC",
				},
			},
			htype: BINARY_TBL,
			table: [][]complex64{
				{complex(float32(10), float32(10)), complex(float32(11), float32(11)),
					complex(float32(12), float32(12)), complex(float32(13), float32(13))},
				{complex(float32(14), float32(14)), complex(float32(15), float32(15)),
					complex(float32(16), float32(16)), complex(float32(17), float32(17))},
				{complex(float32(18), float32(18)), complex(float32(19), float32(19)),
					complex(float32(10), float32(10)), complex(float32(11), float32(11))},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "cplx128s",
					Format: "QM",
				},
			},
			htype: BINARY_TBL,
			table: [][]complex128{
				{complex(10, 10), complex(11, 11), complex(12, 12), complex(13, 13)},
				{complex(14, 14), complex(15, 15), complex(16, 16), complex(17, 17)},
				{complex(18, 18), complex(19, 19), complex(10, 10), complex(11, 11)},
			},
		},
	} {
		fname := fmt.Sprintf("%03d_%s", ii, table.name)
		for _, fct := range []func(){
			// create
			func() {
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

				phdu, err := NewPrimaryHDU(nil)
				if err != nil {
					t.Fatalf("error creating PHDU: %v", err)
				}
				defer phdu.Close()

				err = f.Write(phdu)
				if err != nil {
					t.Fatalf("error writing PHDU: %v", err)
				}

				tbl, err := NewTable("test", table.cols, table.htype)
				if err != nil {
					t.Fatalf("error creating new table: %v (%v)", err, table.cols[0].Name)
				}
				defer tbl.Close()

				rslice := reflect.ValueOf(table.table)
				for i := 0; i < rslice.Len(); i++ {
					data := rslice.Index(i).Addr()
					err = tbl.Write(data.Interface())
					if err != nil {
						t.Fatalf("error writing row [%v]: %v", i, err)
					}
				}

				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v] (%v)", rslice.Len(), nrows, table.cols[0].Name)
				}

				err = f.Write(tbl)
				if err != nil {
					t.Fatalf("error writing table: %v", err)
				}
			},
			// read
			func() {
				r, err := os.Open(fname)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer r.Close()
				f, err := Open(r)
				if err != nil {
					b, _ := ioutil.ReadFile(fname)
					t.Fatalf("error opening file [%v]: %v\n%q", fname, err, string(b))
				}
				defer f.Close()

				hdu := f.HDU(1)
				tbl := hdu.(*Table)
				if tbl.Name() != "test" {
					t.Fatalf("expected table name==%q. got %q", "test", tbl.Name())
				}

				rslice := reflect.ValueOf(table.table)
				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v]", rslice.Len(), nrows)
				}

				rows, err := tbl.Read(0, nrows)
				if err != nil {
					t.Fatalf("table.Read: %v", err)
				}
				count := int64(0)
				for rows.Next() {
					ref := rslice.Index(int(count)).Interface()
					rt := reflect.TypeOf(ref)
					ptr := reflect.New(rt)
					err = rows.Scan(ptr.Interface())
					if err != nil {
						t.Fatalf("rows.Scan: %v", err)
					}
					data := ptr.Elem().Interface()
					// check data just read in is ok
					if !reflect.DeepEqual(data, ref) {
						t.Fatalf("rows.Scan: [%[3]s|%[4]v]\nexp=%[1]v (%[1]T)\ngot=%[2]v (%[2]T)", ref, data, table.cols[0].Name, table.htype)
					}
					count++
				}
				if count != nrows {
					t.Fatalf("expected [%v] rows. got [%v]", nrows, count)
				}
			},
		} {
			fct()
		}
	}
}

func TestTableArrayRW(t *testing.T) {

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

	for ii, table := range []struct {
		name  string
		cols  []Column
		htype HDUType
		table interface{}
	}{
		// binary table
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "bools",
					Format: "4L",
				},
			},
			htype: BINARY_TBL,
			table: [][4]bool{
				{true, true, true, true},
				{false, false, false, false},
				{true, false, true, false},
				{false, true, false, true},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int8s",
					Format: "4B",
				},
			},
			htype: BINARY_TBL,
			table: [][4]int8{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int16s",
					Format: "4I",
				},
			},
			htype: BINARY_TBL,
			table: [][4]int16{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int32s",
					Format: "4J",
				},
			},
			htype: BINARY_TBL,
			table: [][4]int32{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int64s",
					Format: "4K",
				},
			},
			htype: BINARY_TBL,
			table: [][4]int64{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "ints",
					Format: "4K",
				},
			},
			htype: BINARY_TBL,
			table: [][4]int{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint8s",
					Format: "4B",
				},
			},
			htype: BINARY_TBL,
			table: [][4]uint8{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint16s",
					Format: "4I",
				},
			},
			htype: BINARY_TBL,
			table: [][4]uint16{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint32s",
					Format: "4J",
				},
			},
			htype: BINARY_TBL,
			table: [][4]uint32{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uint64s",
					Format: "4K",
				},
			},
			htype: BINARY_TBL,
			table: [][4]uint64{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "uints",
					Format: "4K",
				},
			},
			htype: BINARY_TBL,
			table: [][4]uint{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float32s",
					Format: "4E",
				},
			},
			htype: BINARY_TBL,
			table: [][4]float32{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "float64s",
					Format: "4D",
				},
			},
			htype: BINARY_TBL,
			table: [][4]float64{
				{10, 11, 12, 13},
				{14, 15, 16, 17},
				{18, 19, 10, 11},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "cplx64s",
					Format: "4C",
				},
			},
			htype: BINARY_TBL,
			table: [][4]complex64{
				{complex(float32(10), float32(10)), complex(float32(11), float32(11)),
					complex(float32(12), float32(12)), complex(float32(13), float32(13))},
				{complex(float32(14), float32(14)), complex(float32(15), float32(15)),
					complex(float32(16), float32(16)), complex(float32(17), float32(17))},
				{complex(float32(18), float32(18)), complex(float32(19), float32(19)),
					complex(float32(10), float32(10)), complex(float32(11), float32(11))},
			},
		},
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "cplx128s",
					Format: "4M",
				},
			},
			htype: BINARY_TBL,
			table: [][4]complex128{
				{complex(10, 10), complex(11, 11), complex(12, 12), complex(13, 13)},
				{complex(14, 14), complex(15, 15), complex(16, 16), complex(17, 17)},
				{complex(18, 18), complex(19, 19), complex(10, 10), complex(11, 11)},
			},
		},
	} {
		fname := fmt.Sprintf("%03d_%s", ii, table.name)
		for _, fct := range []func(){
			// create
			func() {
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

				phdu, err := NewPrimaryHDU(nil)
				if err != nil {
					t.Fatalf("error creating PHDU: %v", err)
				}
				defer phdu.Close()

				defer phdu.Close()

				err = f.Write(phdu)
				if err != nil {
					t.Fatalf("error writing PHDU: %v", err)
				}

				tbl, err := NewTable("test", table.cols, table.htype)
				if err != nil {
					t.Fatalf("error creating new table: %v (%v)", err, table.cols[0].Name)
				}
				defer tbl.Close()

				rslice := reflect.ValueOf(table.table)
				for i := 0; i < rslice.Len(); i++ {
					data := rslice.Index(i).Addr()
					err = tbl.Write(data.Interface())
					if err != nil {
						t.Fatalf("error writing row [%v]: %v", i, err)
					}
				}

				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v] (%v)", rslice.Len(), nrows, table.cols[0].Name)
				}

				err = f.Write(tbl)
				if err != nil {
					t.Fatalf("error writing table: %v", err)
				}
			},
			// read
			func() {
				r, err := os.Open(fname)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer r.Close()

				f, err := Open(r)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer f.Close()

				hdu := f.HDU(1)
				tbl := hdu.(*Table)
				if tbl.Name() != "test" {
					t.Fatalf("expected table name==%q. got %q", "test", tbl.Name())
				}

				rslice := reflect.ValueOf(table.table)
				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v]", rslice.Len(), nrows)
				}

				rows, err := tbl.Read(0, nrows)
				if err != nil {
					t.Fatalf("table.Read: %v", err)
				}
				count := int64(0)
				for rows.Next() {
					ref := rslice.Index(int(count)).Interface()
					rt := reflect.TypeOf(ref)
					ptr := reflect.New(rt)
					err = rows.Scan(ptr.Interface())
					if err != nil {
						t.Fatalf("rows.Scan: %v", err)
					}
					data := ptr.Elem().Interface()
					// check data just read in is ok
					if !reflect.DeepEqual(data, ref) {
						t.Fatalf("rows.Scan: [%[3]s|%[4]v]\nexp=%[1]v (%[1]T)\ngot=%[2]v (%[2]T)", ref, data, table.cols[0].Name, table.htype)
					}
					count++
				}
				if count != nrows {
					t.Fatalf("expected [%v] rows. got [%v]", nrows, count)
				}
			},
		} {
			fct()
		}
	}
}

func TestTableStructsRW(t *testing.T) {

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

	type DataStruct struct {
		A   int64      `fits:"int64"`
		XX0 int        // hole
		B   float64    `fits:"float64"`
		XX1 int        // hole
		C   []int64    `fits:"int64s"`
		XX2 int        // hole
		D   []float64  `fits:"float64s"`
		XX3 int        // hole
		E   [2]float64 `fits:"arrfloat64s"`
	}

	for ii, table := range []struct {
		name  string
		cols  []Column
		htype HDUType
		table interface{}
	}{
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "int64",
					Format: "K",
				},
				{
					Name:   "float64",
					Format: "D",
				},
				{
					Name:   "int64s",
					Format: "QK",
				},
				{
					Name:   "float64s",
					Format: "QD",
				},
				{
					Name:   "arrfloat64s",
					Format: "4D",
				},
			},
			htype: BINARY_TBL,
			table: []DataStruct{
				{A: 10, B: 10, C: []int64{10, 10}, D: []float64{10, 10}, E: [2]float64{10, 10}},
				{A: 11, B: 11, C: []int64{11, 11}, D: []float64{11, 11}, E: [2]float64{11, 11}},
				{A: 12, B: 12, C: []int64{12, 12}, D: []float64{12, 12}, E: [2]float64{12, 12}},
				{A: 13, B: 13, C: []int64{13, 13}, D: []float64{13, 13}, E: [2]float64{13, 13}},
			},
		},
	} {
		fname := fmt.Sprintf("%03d_%s", ii, table.name)
		for _, fct := range []func(){
			// create
			func() {
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

				phdu, err := NewPrimaryHDU(nil)
				if err != nil {
					t.Fatalf("error creating PHDU: %v", err)
				}
				defer phdu.Close()

				err = f.Write(phdu)
				if err != nil {
					t.Fatalf("error writing PHDU: %v", err)
				}

				tbl, err := NewTable("test", table.cols, table.htype)
				if err != nil {
					t.Fatalf("error creating new table: %v (%v)", err, table.cols[0].Name)
				}
				defer tbl.Close()

				rslice := reflect.ValueOf(table.table)
				for i := 0; i < rslice.Len(); i++ {
					data := rslice.Index(i).Addr()
					err = tbl.Write(data.Interface())
					if err != nil {
						t.Fatalf("error writing row [%v]: %v", i, err)
					}
				}

				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v] (%v)", rslice.Len(), nrows, table.cols[0].Name)
				}

				err = f.Write(tbl)
				if err != nil {
					t.Fatalf("error writing table: %v", err)
				}
			},
			// read
			func() {
				r, err := os.Open(fname)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer r.Close()
				f, err := Open(r)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer f.Close()

				hdu := f.HDU(1)
				tbl := hdu.(*Table)
				if tbl.Name() != "test" {
					t.Fatalf("expected table name==%q. got %q", "test", tbl.Name())
				}

				rslice := reflect.ValueOf(table.table)
				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v]", rslice.Len(), nrows)
				}

				rows, err := tbl.Read(0, nrows)
				if err != nil {
					t.Fatalf("table.Read: %v", err)
				}
				count := int64(0)
				for rows.Next() {
					ref := rslice.Index(int(count)).Interface()
					rt := reflect.TypeOf(ref)
					rv := reflect.New(rt).Elem()
					data := rv.Interface().(DataStruct)
					err = rows.Scan(&data)
					if err != nil {
						t.Fatalf("rows.Scan: %v", err)
					}
					// check data just read in is ok
					if !reflect.DeepEqual(data, ref) {
						t.Fatalf("rows.Scan:\nexpected=%v\ngot=%v (%T)", ref, data, data)
					}
					count++
				}
				if count != nrows {
					t.Fatalf("expected [%v] rows. got [%v]", nrows, count)
				}
			},
		} {
			fct()
		}
	}
}

func TestTableMapsRW(t *testing.T) {

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

	for ii, table := range []struct {
		name  string
		cols  []Column
		htype HDUType
		table interface{}
	}{
		{
			name: "new.fits",
			cols: []Column{
				{
					Name:   "A",
					Format: "K",
				},
				{
					Name:   "B",
					Format: "D",
				},
				{
					Name:   "C",
					Format: "QK",
				},
				{
					Name:   "D",
					Format: "QD",
				},
				{
					Name:   "E",
					Format: "2D",
				},
			},
			htype: BINARY_TBL,
			table: []map[string]interface{}{
				{
					"A": int64(10),
					"B": float64(10),
					"C": []int64{10, 10},
					"D": []float64{10, 10},
					"E": [2]float64{10, 10},
				},
				{
					"A": int64(11),
					"B": float64(11),
					"C": []int64{11, 11},
					"D": []float64{11, 11},
					"E": [2]float64{11, 11},
				},
				{
					"A": int64(12),
					"B": float64(12),
					"C": []int64{12, 12},
					"D": []float64{12, 12},
					"E": [2]float64{12, 12},
				},
				{
					"A": int64(13),
					"B": float64(13),
					"C": []int64{13, 13},
					"D": []float64{13, 13},
					"E": [2]float64{13, 13},
				},
			},
		},
	} {
		fname := fmt.Sprintf("%03d_%s", ii, table.name)
		for _, fct := range []func(){
			// create
			func() {
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

				phdu, err := NewPrimaryHDU(nil)
				if err != nil {
					t.Fatalf("error creating PHDU: %v", err)
				}
				defer phdu.Close()

				err = f.Write(phdu)
				if err != nil {
					t.Fatalf("error writing PHDU: %v", err)
				}

				tbl, err := NewTable("test", table.cols, table.htype)
				if err != nil {
					t.Fatalf("error creating new table: %v (%v)", err, table.cols[0].Name)
				}
				defer tbl.Close()

				rslice := reflect.ValueOf(table.table)
				for i := 0; i < rslice.Len(); i++ {
					data := rslice.Index(i).Addr()
					err = tbl.Write(data.Interface())
					if err != nil {
						t.Fatalf("error writing row [%v]: %v", i, err)
					}
				}

				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v] (%v)", rslice.Len(), nrows, table.cols[0].Name)
				}

				err = f.Write(tbl)
				if err != nil {
					t.Fatalf("error writing table: %v", err)
				}
			},
			// read
			func() {
				r, err := os.Open(fname)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer r.Close()
				f, err := Open(r)
				if err != nil {
					t.Fatalf("error opening file [%v]: %v", fname, err)
				}
				defer f.Close()

				hdu := f.HDU(1)
				tbl := hdu.(*Table)
				if tbl.Name() != "test" {
					t.Fatalf("expected table name==%q. got %q", "test", tbl.Name())
				}

				rslice := reflect.ValueOf(table.table)
				nrows := tbl.NumRows()
				if nrows != int64(rslice.Len()) {
					t.Fatalf("expected num rows [%v]. got [%v]", rslice.Len(), nrows)
				}

				rows, err := tbl.Read(0, nrows)
				if err != nil {
					t.Fatalf("table.Read: %v", err)
				}
				count := int64(0)
				for rows.Next() {
					ref := rslice.Index(int(count)).Interface().(map[string]interface{})
					data := map[string]interface{}{}
					err = rows.Scan(&data)
					if err != nil {
						t.Fatalf("rows.Scan: %v", err)
					}
					// check data just read in is ok
					if !reflect.DeepEqual(data, ref) {
						t.Fatalf("rows.Scan:\nexp=%[1]v (%[1]T)\ngot=%[2]v (%[2]T)", ref, data)
					}
					count++
				}
				if count != nrows {
					t.Fatalf("expected [%v] rows. got [%v]", nrows, count)
				}
			},
		} {
			fct()
		}
	}
}

// EOF
