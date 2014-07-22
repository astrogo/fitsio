package fitsio

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gonuts/binary"
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
		buf := bytes.NewBuffer(row)
		bdec := binary.NewDecoder(buf)
		bdec.Order = binary.BigEndian

		slice := reflect.ValueOf(ptr).Elem()
		nmax := 0

		switch col.dtype.dsize {
		case 8:
			var n int32
			var offset int32
			err = bdec.Decode(&n)
			if err != nil {
				return fmt.Errorf("fitsio: problem decoding slice 32b-length: %v\n", err)
			}
			err = bdec.Decode(&offset)
			if err != nil {
				return fmt.Errorf("fitsio: problem decoding slice 32b-offset: %v\n", err)
			}
			beg = int(offset)
			end = beg + int(n)*int(col.dtype.gotype.Elem().Size())
			nmax = int(n)

		case 16:
			var n int64
			var offset int64
			err = bdec.Decode(&n)
			if err != nil {
				return fmt.Errorf("fitsio: problem decoding slice 64b-length: %v\n", err)
			}
			err = bdec.Decode(&offset)
			if err != nil {
				return fmt.Errorf("fitsio: problem decoding slice 64b-offset: %v\n", err)
			}
			beg = int(offset)
			end = beg + int(n)*int(col.dtype.gotype.Elem().Size())
			nmax = int(n)
		}

		buf = bytes.NewBuffer(table.heap[beg:end])
		bdec = binary.NewDecoder(buf)
		bdec.Order = binary.BigEndian

		slice.SetLen(0)
		for i := 0; i < nmax; i++ {
			vv := reflect.New(rt.Elem())
			err = bdec.Decode(vv.Interface())
			if err != nil {
				return fmt.Errorf("fitsio: problem encoding: %v", err)
			}
			slice = reflect.Append(slice, vv.Elem())
		}
		if err != nil {
			return fmt.Errorf("fitsio: %v\n", err)
		}
		rv := reflect.ValueOf(ptr)
		rv.Elem().Set(slice)

	case reflect.Array:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + (col.dtype.dsize * col.dtype.len)
		row := table.data[beg:end]
		buf := bytes.NewBuffer(row)
		bdec := binary.NewDecoder(buf)
		bdec.Order = binary.BigEndian

		err = bdec.Decode(ptr)
		if err != nil {
			return fmt.Errorf("fitsio: %v\n", err)
		}

	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:

		//scalar := true
		//err = col.decode(table, n, rt, rv, scalar)

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize
		row := table.data[beg:end]
		buf := bytes.NewBuffer(row)
		bdec := binary.NewDecoder(buf)
		bdec.Order = binary.BigEndian
		err = bdec.Decode(ptr)
		if err != nil {
			return fmt.Errorf("fitsio: %v\n", err)
		}

	case reflect.String:

		//scalar := true
		//err = col.decode(table, n, rt, rv, scalar)

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
	var err error

	rv := reflect.Indirect(reflect.ValueOf(ptr))
	rt := reflect.TypeOf(rv.Interface())

	switch rt.Kind() {
	case reflect.Slice:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		buf := &sectionWriter{
			buf: table.data[beg:end],
			beg: 0,
		}
		benc := binary.NewEncoder(buf)
		benc.Order = binary.BigEndian

		slice := reflect.ValueOf(ptr).Elem()
		nmax := slice.Len()

		switch col.dtype.dsize {
		case 8:
			data := [2]int32{int32(nmax), int32(len(table.heap))}
			err = benc.Encode(&data)
			if err != nil {
				return fmt.Errorf("fitsio: problem encoding slice 32b-descriptor: %v\n", err)
			}

		case 16:
			data := [2]int64{int64(nmax), int64(len(table.heap))}
			err = benc.Encode(&data)
			if err != nil {
				return fmt.Errorf("fitsio: problem encoding slice 64b-descriptor: %v\n", err)
			}
		}

		{
			buf := new(bytes.Buffer)
			buf.Grow(nmax * col.dtype.hsize)
			benc = binary.NewEncoder(buf)
			benc.Order = binary.BigEndian

			for i := 0; i < nmax; i++ {
				err = benc.Encode(slice.Index(i).Addr().Interface())
				if err != nil {
					return fmt.Errorf("fitsio: problem encoding: %v", err)
				}
			}
			table.heap = append(table.heap, buf.Bytes()...)
		}

	case reflect.Array:

		beg := table.rowsz*int(irow) + col.offset
		end := beg + (col.dtype.dsize * col.dtype.len)

		//buf := new(bytes.Buffer)
		buf := &sectionWriter{
			buf: table.data[beg:end],
			beg: 0,
		}
		benc := binary.NewEncoder(buf)
		benc.Order = binary.BigEndian
		err = benc.Encode(ptr)
		if err != nil {
			return fmt.Errorf("fitsio: %v\n", err)
		}

	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:

		//scalar := true
		//err = col.decode(table, n, rt, rv, scalar)

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		//buf := bytes.NewBuffer(row)
		buf := &sectionWriter{
			buf: table.data[beg:end],
			beg: 0,
		}

		benc := binary.NewEncoder(buf)
		benc.Order = binary.BigEndian
		err = benc.Encode(ptr)

	case reflect.String:

		//scalar := true
		//err = col.decode(table, n, rt, rv, scalar)

		beg := table.rowsz*int(irow) + col.offset
		end := beg + col.dtype.dsize

		//buf := bytes.NewBuffer(row)
		buf := &sectionWriter{
			buf: table.data[beg:end],
			beg: 0,
		}

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
	buf := &sectionWriter{
		buf: table.data[beg:end],
		beg: 0,
	}

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
		n, err = fmt.Fprintf(buf, str)
		if err != nil {
			return fmt.Errorf("fitsio: error writing '%#v': %v", rv.Interface(), err)
		}
		if n != len(buf.buf) {
			return fmt.Errorf(
				"fitsio: error writing '%#v'. expected %d bytes. wrote %d",
				rv.Interface(), len(buf.buf), n,
			)
		}

	default:
		return fmt.Errorf("fitsio: ASCII-table can not read/write %v", rt.Kind())
	}

	return err
}
