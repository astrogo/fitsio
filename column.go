// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Column represents a column in a FITS table
type Column struct {
	Name    string  // column name, corresponding to ``TTYPE`` keyword
	Format  string  // column format, corresponding to ``TFORM`` keyword
	Unit    string  // column unit, corresponding to ``TUNIT`` keyword
	Null    string  // null value, corresponding to ``TNULL`` keyword
	Bscale  float64 // bscale value, corresponding to ``TSCAL`` keyword
	Bzero   float64 // bzero value, corresponding to ``TZERO`` keyword
	Display string  // display format, corresponding to ``TDISP`` keyword
	Dim     []int64 // column dimension corresponding to ``TDIM`` keyword
	Start   int64   // column starting position, corresponding to ``TBCOL`` keyword

	dtype  Type   // type of the value held by the column
	offset int    // offset in bytes (from start of data-pad or heap-pad) to get at the column data
	txtfmt string // go-based fmt string for ASCII table

	// read function (binary/ascii)
	read func(table *Table, icol int, irow int64, ptr interface{}) error

	// write function (binary/ascii)
	write func(table *Table, icol int, irow int64, ptr interface{}) error
}

// NewColumn creates a new Column with name `name` and Format inferred from the type of value
func NewColumn(name string, v interface{}) (Column, error) {
	panic("not implemented")
}

// Type returns the Go reflect.Type associated with this Column
func (col *Column) Type() reflect.Type {
	return col.dtype.gotype
}

// readBin reads the value at column number icol and row irow, into ptr.
func (col *Column) readBin(table *Table, icol int, irow int64, ptr interface{}) error {
	var err error

	rv := reflect.Indirect(reflect.ValueOf(ptr))
	rt := reflect.TypeOf(rv.Interface())

	switch rt.Kind() {
	case reflect.Slice:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize
		row := table.data[beg:end]
		r := newReader(row)
		slice := reflect.ValueOf(ptr).Elem()
		nmax := 0

		switch col.dtype.dsize {
		case 8:
			var n int32
			var offset int32
			r.readI32(&n)
			r.readI32(&offset)
			beg = int(offset)
			end = beg + int(n)*int(col.dtype.gotype.Elem().Size())
			nmax = int(n)

		case 16:
			var n int64
			var offset int64
			r.readI64(&n)
			r.readI64(&offset)
			beg = int(offset)
			end = beg + int(n)*int(col.dtype.gotype.Elem().Size())
			nmax = int(n)
		}
		if slice.Len() < nmax {
			slice = reflect.MakeSlice(rt, nmax, nmax)
		}

		r = newReader(table.heap[beg:end])
		switch slice := slice.Interface().(type) {
		case []bool:
			r.readBools(slice[:nmax])

		case []byte:
			copy(slice[:nmax], r.p)

		case []int8:
			r.readI8s(slice[:nmax])

		case []int16:
			r.readI16s(slice[:nmax])

		case []int32:
			r.readI32s(slice[:nmax])

		case []int64:
			r.readI64s(slice[:nmax])

		case []int:
			r.readInts(slice[:nmax])

		case []uint16:
			r.readU16s(slice[:nmax])

		case []uint32:
			r.readU32s(slice[:nmax])

		case []uint64:
			r.readU64s(slice[:nmax])

		case []uint:
			r.readUints(slice[:nmax])

		case []float32:
			r.readF32s(slice[:nmax])

		case []float64:
			r.readF64s(slice[:nmax])

		case []complex64:
			r.readC64s(slice[:nmax])

		case []complex128:
			r.readC128s(slice[:nmax])

		default:
			panic(fmt.Errorf("fitsio: not implemented %T", slice))
		}
		rv.Set(slice)

	case reflect.Array:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + (col.dtype.dsize * col.dtype.len)
		row := table.data[beg:end]
		r := newReader(row)
		switch slice := rv.Slice(0, rv.Len()).Interface().(type) {
		case []bool:
			r.readBools(slice)

		case []byte:
			copy(slice, r.p)

		case []int8:
			r.readI8s(slice)

		case []int16:
			r.readI16s(slice)

		case []int32:
			r.readI32s(slice)

		case []int64:
			r.readI64s(slice)

		case []int:
			r.readInts(slice)

		case []uint16:
			r.readU16s(slice)

		case []uint32:
			r.readU32s(slice)

		case []uint64:
			r.readU64s(slice)

		case []uint:
			r.readUints(slice)

		case []float32:
			r.readF32s(slice)

		case []float64:
			r.readF64s(slice)

		case []complex64:
			r.readC64s(slice)

		case []complex128:
			r.readC128s(slice)

		default:
			panic(fmt.Errorf("fitsio: not implemented %T", slice))
		}

	case reflect.Bool:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readBool(ptr.(*bool))

	case reflect.Int8:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readI8(ptr.(*int8))

	case reflect.Int16:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readI16(ptr.(*int16))

	case reflect.Int32:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readI32(ptr.(*int32))

	case reflect.Int64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readI64(ptr.(*int64))

	case reflect.Int:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readInt(ptr.(*int))

	case reflect.Uint8:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readU8(ptr.(*uint8))

	case reflect.Uint16:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readU16(ptr.(*uint16))

	case reflect.Uint32:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readU32(ptr.(*uint32))

	case reflect.Uint64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readU64(ptr.(*uint64))

	case reflect.Uint:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readUint(ptr.(*uint))

	case reflect.Float32:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readF32(ptr.(*float32))

	case reflect.Float64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readF64(ptr.(*float64))

	case reflect.Complex64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readC64(ptr.(*complex64))

	case reflect.Complex128:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		r := newReader(table.data[beg:end])
		r.readC128(ptr.(*complex128))

	case reflect.String:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize
		row := table.data[beg:end]
		str := ""
		if row[0] == '\x00' {
			str = string(row[1:])
			str = strings.TrimRight(str, string([]byte("\x00")))
		} else {
			str = string(row)
		}

		rv.SetString(str)

	default:
		return fmt.Errorf("fitsio: binary-table can not read/write %v", rt.Kind())
	}
	return err
}

// writeBin writes the value at column number icol and row irow, from ptr.
func (col *Column) writeBin(table *Table, icol int, irow int64, ptr interface{}) error {
	var (
		err error
		rv  = reflect.Indirect(reflect.ValueOf(ptr))
		rvi = rv.Interface()
		rt  = reflect.TypeOf(rvi)
	)

	switch rt.Kind() {
	case reflect.Slice:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		nmax := rv.Len()

		switch col.dtype.dsize {
		case 8:
			w.writeI32(int32(nmax))
			w.writeI32(int32(len(table.heap)))

		case 16:
			w.writeI64(int64(nmax))
			w.writeI64(int64(len(table.heap)))
		}

		{
			w := newWriter(make([]byte, nmax*col.dtype.hsize))
			switch slice := rvi.(type) {
			case []bool:
				w.writeBools(slice)

			case []byte:
				copy(w.p, slice)

			case []int8:
				w.writeI8s(slice)

			case []int16:
				w.writeI16s(slice)

			case []int32:
				w.writeI32s(slice)

			case []int64:
				w.writeI64s(slice)

			case []int:
				w.writeInts(slice)

			case []uint16:
				w.writeU16s(slice)

			case []uint32:
				w.writeU32s(slice)

			case []uint64:
				w.writeU64s(slice)

			case []uint:
				w.writeUints(slice)

			case []float32:
				w.writeF32s(slice)

			case []float64:
				w.writeF64s(slice)

			case []complex64:
				w.writeC64s(slice)

			case []complex128:
				w.writeC128s(slice)

			default:
				panic(fmt.Errorf("fitsio: not implemented %T", slice))
			}
			table.heap = append(table.heap, w.bytes()...)
		}

	case reflect.Array:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + (col.dtype.dsize * col.dtype.len)

		w := newWriter(table.data[beg:end])
		switch slice := rv.Slice(0, rv.Len()).Interface().(type) {
		case []bool:
			w.writeBools(slice)

		case []byte:
			copy(w.p, slice)

		case []int8:
			w.writeI8s(slice)

		case []int16:
			w.writeI16s(slice)

		case []int32:
			w.writeI32s(slice)

		case []int64:
			w.writeI64s(slice)

		case []int:
			w.writeInts(slice)

		case []uint16:
			w.writeU16s(slice)

		case []uint32:
			w.writeU32s(slice)

		case []uint64:
			w.writeU64s(slice)

		case []uint:
			w.writeUints(slice)

		case []float32:
			w.writeF32s(slice)

		case []float64:
			w.writeF64s(slice)

		case []complex64:
			w.writeC64s(slice)

		case []complex128:
			w.writeC128s(slice)

		default:
			panic(fmt.Errorf("fitsio: not implemented %T", slice))
		}

	case reflect.Bool:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeBool(rvi.(bool))

	case reflect.Int8:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeI8(rvi.(int8))

	case reflect.Int16:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeI16(rvi.(int16))

	case reflect.Int32:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeI32(rvi.(int32))

	case reflect.Int64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeI64(rvi.(int64))

	case reflect.Int:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeInt(rvi.(int))

	case reflect.Uint8:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeU8(rvi.(uint8))

	case reflect.Uint16:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeU16(rvi.(uint16))

	case reflect.Uint32:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeU32(rvi.(uint32))

	case reflect.Uint64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeU64(rvi.(uint64))

	case reflect.Uint:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeUint(rvi.(uint))

	case reflect.Float32:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeF32(rvi.(float32))

	case reflect.Float64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeF64(rvi.(float64))

	case reflect.Complex64:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeC64(rvi.(complex64))

	case reflect.Complex128:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		w := newWriter(table.data[beg:end])
		w.writeC128(rvi.(complex128))

	case reflect.String:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		buf := newWriter(table.data[beg:end])
		str := rv.String()
		data := make([]byte, 0, len(str)+1)
		data = append(data, '\x00')
		data = append(data, []byte(str)...)
		n := len(data)
		if n > end-beg {
			n = end - beg
		}
		if n < end-beg {
			data = append(data, bytes.Repeat([]byte("\x00"), end-beg-n)...)
			n = len(data)
		}
		_, err = buf.Write(data[:n])

		if err != nil {
			return fmt.Errorf("fitsio: %v\n", err)
		}

	default:
		return fmt.Errorf("fitsio: binary-table can not read/write %v", rt.Kind())
	}

	return err
}

// readTxt reads the value at column number icol and row irow, into ptr.
func (col *Column) readTxt(table *Table, icol int, irow int64, ptr interface{}) error {
	var err error

	rv := reflect.Indirect(reflect.ValueOf(ptr))
	rt := reflect.TypeOf(rv.Interface())

	beg := table.rowsz*int(irow) + col.offset
	end := beg + col.dtype.dsize
	buf := table.data[beg:end]
	str := strings.TrimSpace(string(buf))

	switch rt.Kind() {
	case reflect.Slice:

		return fmt.Errorf("fitsio: ASCII-table can not read/write slices")

	case reflect.Array:

		return fmt.Errorf("fitsio: ASCII-table can not read/write arrays")

	case reflect.Bool:

		return fmt.Errorf("fitsio: ASCII-table can not read/write booleans")

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("fitsio: error parsing %q into a uint: %v", str, err)
		}
		rv.SetUint(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("fitsio: error parsing %q into an int: %v", str, err)
		}
		rv.SetInt(v)

	case reflect.Float32, reflect.Float64:

		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("fitsio: error parsing %q into a float: %v", str, err)
		}
		rv.SetFloat(v)

	case reflect.Complex64, reflect.Complex128:

		return fmt.Errorf("fitsio: ASCII-table can not read/write complexes")

	case reflect.String:
		rv.SetString(str)

	default:
		return fmt.Errorf("fitsio: ASCII-table can not read/write %v", rt.Kind())
	}
	return err
}

// writeTxt writes the value at column number icol and row irow, from ptr.
func (col *Column) writeTxt(table *Table, icol int, irow int64, ptr interface{}) error {
	var err error

	beg := table.rowsz*int(irow) + col.offset
	end := beg + col.dtype.dsize
	w := newWriter(table.data[beg:end])

	rv := reflect.Indirect(reflect.ValueOf(ptr))
	rt := reflect.TypeOf(rv.Interface())

	switch rt.Kind() {
	case reflect.Slice:

		return fmt.Errorf("fitsio: ASCII-table can not read/write slices")

	case reflect.Array:

		return fmt.Errorf("fitsio: ASCII-table can not read/write arrays")

	case reflect.Bool:

		return fmt.Errorf("fitsio: ASCII-table can not read/write booleans")

	case reflect.Complex64, reflect.Complex128:

		return fmt.Errorf("fitsio: ASCII-table can not read/write complexes")

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64,
		reflect.String:

		str := fmt.Sprintf(col.txtfmt, rv.Interface())
		n := 0
		n, err = fmt.Fprintf(w, str)
		if err != nil {
			return fmt.Errorf("fitsio: error writing '%#v': %v", rv.Interface(), err)
		}
		if n != len(w.p) {
			return fmt.Errorf(
				"fitsio: error writing '%#v'. expected %d bytes. wrote %d",
				rv.Interface(), len(w.p), n,
			)
		}

	default:
		return fmt.Errorf("fitsio: ASCII-table can not read/write %v", rt.Kind())
	}

	return err
}
