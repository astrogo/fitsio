// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
	"reflect"
)

// Rows is the result of a query on a FITS Table.
// Its cursors starts before the first row of the result set.
// Use Next to advance through the rows:
//
//  rows, err := table.Read(0, -1)
//  ...
//  for rows.Next() {
//      var id int
//      var x float64
//      err = rows.Scan(&id, &x)
//      ...
//  }
//  err = rows.Err() // get any error encountered during iteration
//  ...
//
type Rows struct {
	table  *Table
	cols   []int // list of (active) column indices
	i      int64 // number of rows iterated over
	n      int64 // number of rows this iterator iters over
	inc    int64 // number of rows to increment by at each iteration
	cur    int64 // current row index
	closed bool
	err    error // last error

	// cache of type -> slice of (struct-field-index,col-index)
	// used by scanStruct
	icols map[reflect.Type][][2]int
}

// Err returns the error, if any, that was encountered during iteration.
// Err may be called after an explicit or implicit Close.
func (rows *Rows) Err() error {
	return rows.err
}

// Close closes the Rows, preventing further enumeration.
// Close is idempotent and does not affect the result of Err.
func (rows *Rows) Close() error {
	if rows.closed {
		return nil
	}
	rows.closed = true
	rows.table = nil
	return nil
}

// Scan copies the columns in the current row into the values pointed at by
// dest.
func (rows *Rows) Scan(args ...interface{}) error {
	var err error
	defer func() {
		rows.err = err
	}()

	switch len(args) {
	case 0:
		return fmt.Errorf("fitsio: Rows.Scan needs at least one argument")

	case 1:
		// maybe special case: map? struct?
		rt := reflect.TypeOf(args[0]).Elem()
		switch rt.Kind() {
		case reflect.Map:
			return rows.scanMap(*args[0].(*map[string]interface{}))
		case reflect.Struct:
			err = rows.scanStruct(args[0])
			return err
		}
	}

	return rows.scan(args...)
}

func (rows *Rows) scan(args ...interface{}) error {
	var err error
	if len(args) != len(rows.cols) {
		return fmt.Errorf(
			"fitsio.Rows.Scan: invalid number of arguments (got %d. expected %d)",
			len(args),
			len(rows.cols),
		)
	}
	for i, icol := range rows.cols {
		err = rows.table.cols[icol].read(rows.table, icol, rows.cur, args[i])
		if err != nil {
			return err
		}
	}
	return err
}

func (rows *Rows) scanMap(data map[string]interface{}) error {
	var err error
	icols := make([]int, 0, len(data))
	switch len(data) {
	case 0:
		icols = make([]int, len(rows.cols))
		for i := range rows.cols {
			icols[i] = i
		}
	default:
		for k := range data {
			icol := rows.table.Index(k)
			if icol >= 0 {
				icols = append(icols, icol)
			}
		}
	}

	for _, icol := range icols {
		col := rows.table.Col(icol)
		val := reflect.New(col.Type())
		err = col.read(rows.table, icol, rows.cur, val.Interface())
		if err != nil {
			return err
		}
		data[col.Name] = val.Elem().Interface()
	}
	return err
}

func (rows *Rows) scanStruct(data interface{}) error {
	var err error

	rt := reflect.TypeOf(data).Elem()
	rv := reflect.ValueOf(data).Elem()
	if _, ok := rows.icols[rt]; !ok {
		icols := make([][2]int, 0, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			n := f.Tag.Get("fits")
			if n == "" {
				n = f.Name
			}
			icol := rows.table.Index(n)
			if icol >= 0 {
				icols = append(icols, [2]int{i, icol})
			}
		}
		rows.icols[rt] = icols
	}
	icols := rows.icols[rt]
	for _, icol := range icols {
		col := &rows.table.cols[icol[1]]
		value := rv.Field(icol[0]).Addr().Interface()
		err = col.read(rows.table, icol[1], rows.cur, value)
		if err != nil {
			return err
		}
	}
	return err
}

// Next prepares the next result row for reading with the Scan method.
// It returns true on success, false if there is no next result row.
// Every call to Scan, even the first one, must be preceded by a call to Next.
func (rows *Rows) Next() bool {
	if rows.closed {
		return false
	}
	next := rows.i < rows.n
	rows.cur += rows.inc
	rows.i += rows.inc
	if !next {
		rows.err = rows.Close()
	}
	return next
}
