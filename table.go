// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"fmt"
	"reflect"
)

type Table struct {
	hdr    Header
	binary bool

	data []byte // main data table
	heap []byte // heap data table (for variable length arrays)

	rowsz  int   // size of each row in bytes (ie: NAXIS1)
	nrows  int64 // number of rows (ie: NAXIS2)
	cols   []Column
	colidx map[string]int // associates a column name to its index
}

// Close closes this HDU, cleaning up cycles (if any) for garbage collection
func (t *Table) Close() error {
	return nil
}

// Header returns the Header part of this HDU block.
func (t *Table) Header() *Header {
	return &t.hdr
}

// Type returns the Type of this HDU
func (t *Table) Type() HDUType {
	return t.hdr.Type()
}

// Name returns the value of the 'EXTNAME' Card.
func (t *Table) Name() string {
	card := t.hdr.Get("EXTNAME")
	if card == nil {
		return ""
	}
	return card.Value.(string)
}

// Version returns the value of the 'EXTVER' Card (or 1 if none)
func (t *Table) Version() int {
	card := t.hdr.Get("EXTVER")
	if card == nil {
		return 1
	}
	return card.Value.(int)
}

// Data returns the image payload
func (t *Table) Data() (Value, error) {
	panic("not implemented")
}

func (t *Table) NumRows() int64 {
	return t.nrows
}

func (t *Table) NumCols() int {
	return len(t.cols)
}

func (t *Table) Cols() []Column {
	return t.cols
}

func (t *Table) Col(i int) *Column {
	return &t.cols[i]
}

// Index returns the index of the first column with name `n` or -1
func (t *Table) Index(n string) int {
	idx, ok := t.colidx[n]
	if !ok {
		return -1
	}
	return idx
}

// ReadRange reads rows over the range [beg, end) and returns the corresponding iterator.
// if end > maxrows, the iteration will stop at maxrows
// ReadRange has the same semantics than a `for i=0; i < max; i+=inc {...}` loop
func (t *Table) ReadRange(beg, end, inc int64) (*Rows, error) {
	var err error
	var rows *Rows

	maxrows := t.NumRows()
	if end > maxrows {
		end = maxrows
	}

	if beg < 0 {
		beg = 0
	}

	cols := make([]int, len(t.cols))
	for i := range t.cols {
		cols[i] = i
	}

	rows = &Rows{
		table: t,
		cols:  cols,
		i:     beg,
		n:     end,
		inc:   inc,
		cur:   beg - inc,
		err:   nil,
		icols: make(map[reflect.Type][][2]int),
	}
	return rows, err
}

// Read reads rows over the range [beg, end) and returns the corresponding iterator.
// if end > maxrows, the iteration will stop at maxrows
// ReadRange has the same semantics than a `for i=0; i < max; i++ {...}` loop
func (t *Table) Read(beg, end int64) (*Rows, error) {
	return t.ReadRange(beg, end, 1)
}

// NewTable creates a new table in the given FITS file
func NewTable(name string, cols []Column, hdutype HDUType) (*Table, error) {
	var err error

	isbinary := true
	switch hdutype {
	case ASCII_TBL:
		isbinary = false
	case BINARY_TBL:
		isbinary = true
	default:
		return nil, fmt.Errorf("fitsio: invalid HDUType (%v)", hdutype)
	}

	ncols := len(cols)
	table := &Table{
		hdr:    Header{},
		binary: isbinary,
		data:   make([]byte, 0),
		heap:   make([]byte, 0),
		rowsz:  0, // NAXIS1 in bytes
		nrows:  0, // NAXIS2
		cols:   make([]Column, ncols),
		colidx: make(map[string]int, ncols),
	}

	copy(table.cols, cols)

	cards := make([]Card, 0, len(cols)+2)
	cards = append(
		cards,
		Card{
			Name:    "TFIELDS",
			Value:   ncols,
			Comment: "number of fields in each row",
		},
	)

	offset := 0
	for i := 0; i < ncols; i++ {
		col := &table.cols[i]
		col.offset = offset
		switch hdutype {
		case BINARY_TBL:
			col.write = col.writeBin
			col.read = col.readBin
		case ASCII_TBL:
			col.write = col.writeTxt
			col.read = col.readTxt
		default:
			return nil, fmt.Errorf("fitsio: invalid HDUType (%v)", hdutype)
		}

		table.colidx[col.Name] = i

		if col.Format == "" {
			return nil, fmt.Errorf("fitsio: column (col=%s) has NO valid format", col.Name)
		}

		cards = append(cards,
			Card{
				Name:    fmt.Sprintf("TTYPE%d", i+1),
				Value:   col.Name,
				Comment: fmt.Sprintf("label for column %d", i+1),
			},
			Card{
				Name:    fmt.Sprintf("TFORM%d", i+1),
				Value:   col.Format,
				Comment: fmt.Sprintf("data format for column %d", i+1),
			},
		)

		col.dtype, err = typeFromForm(col.Format, hdutype)
		if err != nil {
			return nil, err
		}

		offset += col.dtype.dsize * col.dtype.len
		col.txtfmt = txtfmtFromForm(col.Format)

		if offset == 0 && i > 0 {
			return nil, fmt.Errorf("fitsio: invalid data-layout")
		}

		if col.Unit != "" {
			cards = append(cards,
				Card{
					Name:    fmt.Sprintf("TUNIT%d", i+1),
					Value:   col.Unit,
					Comment: fmt.Sprintf("unit for column %d", i+1),
				},
			)
		}

		if col.Null != "" {
			cards = append(cards,
				Card{
					Name:    fmt.Sprintf("TNULL%d", i+1),
					Value:   col.Null,
					Comment: fmt.Sprintf("default value for column %d", i+1),
				},
			)
		}

		cards = append(cards,
			Card{
				Name:    fmt.Sprintf("TSCAL%d", i+1),
				Value:   col.Bscale,
				Comment: fmt.Sprintf("scaling offset for column %d", i+1),
			},
		)

		cards = append(cards,
			Card{
				Name:    fmt.Sprintf("TZERO%d", i+1),
				Value:   col.Bzero,
				Comment: fmt.Sprintf("zero value for column %d", i+1),
			},
		)

		if col.Start != 0 {
			cards = append(cards,
				Card{
					Name:  fmt.Sprintf("TBCOL%d", i+1),
					Value: int(col.Start),
				},
			)
		} else {
			cards = append(cards,
				Card{
					Name:  fmt.Sprintf("TBCOL%d", i+1),
					Value: offset - col.dtype.dsize*col.dtype.len + 1,
				},
			)
		}

		if col.Display != "" {
			cards = append(cards,
				Card{
					Name:    fmt.Sprintf("TDISP%d", i+1),
					Value:   col.Display,
					Comment: fmt.Sprintf("display format for column %d", i+1),
				},
			)
		}

		if len(col.Dim) > 0 {
			str := "("
			for idim, dim := range col.Dim {
				str += fmt.Sprintf("%d", dim)
				if idim+1 < len(col.Dim) {
					str += ","
				}
			}
			str += ")"
			cards = append(cards,
				Card{
					Name:  fmt.Sprintf("TDIM%d", i+1),
					Value: str,
				},
			)
		}

	}

	cards = append(
		cards,
		Card{
			Name:    "EXTNAME",
			Value:   name,
			Comment: "name of this table extension",
		},
	)

	bitpix := 8
	hdr := newHeader(cards, hdutype, bitpix, []int{offset, 0})
	table.hdr = *hdr
	table.rowsz = offset

	return table, err
}

// NewTableFrom creates a new table in the given FITS file, using the struct v as schema
func NewTableFrom(name string, v Value, hdutype HDUType) (*Table, error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("fitsio: NewTableFrom takes a struct value. got: %T", v)
	}
	nmax := rt.NumField()
	cols := make([]Column, 0, nmax)
	for i := 0; i < nmax; i++ {
		ft := rt.Field(i)
		name := ft.Tag.Get("fits")
		if name == "" {
			name = ft.Name
		}
		field := rv.Field(i)
		form := formFromGoType(field.Type(), hdutype)
		if form == "" {
			return nil, fmt.Errorf("fitsio: no FITS TFORM for field [%d] %#v", i, field.Interface())
		}
		cols = append(cols,
			Column{
				Name:   name,
				Format: form,
			},
		)
	}
	return NewTable(name, cols, hdutype)
}

// Write writes the data into the columns at the current row.
func (t *Table) Write(args ...interface{}) error {
	var err error

	t.data = append(t.data, make([]byte, t.rowsz)...)

	switch len(args) {
	case 0:
		return fmt.Errorf("fitsio: Rows.Scan needs at least one argument")

	case 1:
		// maybe special case: map? struct?
		rt := reflect.TypeOf(args[0]).Elem()
		switch rt.Kind() {
		case reflect.Map:
			err = t.writeMap(*args[0].(*map[string]interface{}))
		case reflect.Struct:
			err = t.writeStruct(args[0])
		default:
			err = t.write(args[0])
		}
	default:
		err = t.write(args...)
	}

	if err != nil {
		return err
	}

	t.nrows += 1
	t.hdr.axes[1] += 1
	return err
}

func (t *Table) write(args ...interface{}) error {
	var err error
	if len(args) != len(t.cols) {
		return fmt.Errorf(
			"fitsio.Table.Write: invalid number of arguments (got %d. expected %d)",
			len(args),
			len(t.cols),
		)
	}

	for i := range t.cols {
		err = t.cols[i].write(t, i, t.nrows, args[i])
		if err != nil {
			return err
		}
	}

	return err
}

func (t *Table) writeMap(data map[string]interface{}) error {
	var err error
	icols := make([]int, 0, len(data))
	switch len(data) {
	case 0:
		icols = make([]int, len(t.cols))
		for i := range t.cols {
			icols[i] = i
		}
	default:
		for k := range data {
			icol := t.Index(k)
			if icol >= 0 {
				icols = append(icols, icol)
			}
		}
	}

	for _, icol := range icols {
		col := t.Col(icol)
		val := reflect.New(col.Type())
		err = col.write(t, icol, t.nrows, val.Interface())
		if err != nil {
			return err
		}
		data[col.Name] = val.Elem().Interface()
	}
	return err
}

func (t *Table) writeStruct(data interface{}) error {
	var err error
	rt := reflect.TypeOf(data).Elem()
	rv := reflect.ValueOf(data).Elem()
	icols := make([][2]int, 0, rt.NumField())

	if true { // fixme: devise a cache ?
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			n := f.Tag.Get("fits")
			if n == "" {
				n = f.Name
			}
			icol := t.Index(n)
			if icol >= 0 {
				icols = append(icols, [2]int{i, icol})
			}
		}
	}

	for _, icol := range icols {
		col := &t.cols[icol[1]]
		field := rv.Field(icol[0])
		value := field.Addr().Interface()
		err = col.write(t, icol[1], t.nrows, value)
		if err != nil {
			return err
		}
	}
	return err
}

// freeze freezes a Table before writing, calculating offsets and finalizing header values.
func (t *Table) freeze() error {
	var err error
	nrows := t.nrows
	t.hdr.axes[1] = int(nrows)

	if card := t.Header().Get("XTENSION"); card == nil {
		hduext := ""
		if t.binary {
			hduext = "BINTABLE"
		} else {
			hduext = "TABLE   "
		}
		cards := []Card{
			{
				Name:    "XTENSION",
				Value:   hduext,
				Comment: "table extension",
			},
			{
				Name:    "BITPIX",
				Value:   t.Header().Bitpix(),
				Comment: "number of bits per data pixel",
			},
			{
				Name:    "NAXIS",
				Value:   len(t.Header().Axes()),
				Comment: "number of data axes",
			},
			{
				Name:    "NAXIS1",
				Value:   t.Header().Axes()[0],
				Comment: "length of data axis 1",
			},
			{
				Name:    "NAXIS2",
				Value:   t.Header().Axes()[1],
				Comment: "length of data axis 2",
			},
			{
				Name:    "PCOUNT",
				Value:   len(t.heap),
				Comment: "heap area size (bytes)",
			},
			{
				Name:    "GCOUNT",
				Value:   1,
				Comment: "one data group",
			},
		}

		err = t.hdr.prepend(cards...)
		if err != nil {
			return err
		}
	}

	if card := t.Header().Get("THEAP"); card == nil {
		err = t.hdr.Append([]Card{
			{
				Name:    "THEAP",
				Value:   0,
				Comment: "gap size (bytes)",
			},
		}...)
	}

	return err
}

// CopyTable copies all the rows from src into dst.
func CopyTable(dst, src *Table) error {
	return CopyTableRange(dst, src, 0, src.NumRows())
}

// CopyTableRange copies the rows interval [beg,end) from src into dst
func CopyTableRange(dst, src *Table, beg, end int64) error {
	var err error
	if dst == nil {
		return fmt.Errorf("fitsio: dst pointer is nil")
	}
	if src == nil {
		return fmt.Errorf("fitsio: src pointer is nil")
	}

	vla := false
	for _, col := range src.Cols() {
		if col.dtype.tc < 0 {
			vla = true
			break
		}
	}

	// FIXME(sbinet)
	// need to also handle VLAs
	// src.heap -> dst.heap
	// convert offsets into dst.heap
	//
	// for the time being: go the slow way
	switch vla {

	case true:
		rows, err := src.Read(beg, end)
		if err != nil {
			return err
		}
		defer rows.Close()

		ncols := len(src.Cols())
		data := make([]interface{}, ncols)
		for i := range src.Cols() {
			col := &src.cols[i]
			rt := col.dtype.gotype
			rv := reflect.New(rt)
			xx := rv.Interface()
			data[i] = xx
		}
		for rows.Next() {
			err = rows.Scan(data...)
			if err != nil {
				return err
			}
			err = dst.Write(data...)
			if err != nil {
				return err
			}
		}
		err = rows.Err()
		if err != nil {
			return err
		}

		return err

	case false:
		nrows := end - beg
		// reserve enough capacity for the new rows
		dst.data = dst.data[:len(dst.data) : len(dst.data)+int(nrows)*src.rowsz]
		for irow := beg; irow < end; irow++ {
			pstart := src.rowsz * int(irow)
			pend := pstart + src.rowsz
			row := src.data[pstart:pend]
			dst.data = append(dst.data, row...)
		}
		dst.nrows += nrows
		dst.hdr.Axes()[1] += int(nrows)
	}

	return err
}
