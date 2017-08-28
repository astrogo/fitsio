// Copyright 2017 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"encoding/binary"
	"io"
	"math"
)

type rbuf struct {
	p []byte // buffer of data to read from
	c int    // current position in buffer of data
}

func newReader(data []byte) *rbuf {
	return &rbuf{p: data, c: 0}
}

func (r *rbuf) Read(p []byte) (int, error) {
	if r.c >= len(r.p) {
		return 0, io.EOF
	}
	n := copy(p, r.p[r.c:])
	r.c += n
	return n, nil
}

func (r *rbuf) readByte(v *byte) {
	*v = r.p[r.c]
	r.c++
}

func (r *rbuf) readI8(v *int8) {
	*v = int8(r.p[r.c])
	r.c++
}

func (r *rbuf) readI16(v *int16) {
	beg := r.c
	r.c += 2
	*v = int16(binary.BigEndian.Uint16(r.p[beg:r.c]))
}

func (r *rbuf) readI32(v *int32) {
	beg := r.c
	r.c += 4
	*v = int32(binary.BigEndian.Uint32(r.p[beg:r.c]))
}

func (r *rbuf) readI64(v *int64) {
	beg := r.c
	r.c += 8
	*v = int64(binary.BigEndian.Uint64(r.p[beg:r.c]))
}

func (r *rbuf) readF32(v *float32) {
	beg := r.c
	r.c += 4
	*v = math.Float32frombits(binary.BigEndian.Uint32(r.p[beg:r.c]))
}

func (r *rbuf) readF64(v *float64) {
	beg := r.c
	r.c += 8
	*v = math.Float64frombits(binary.BigEndian.Uint64(r.p[beg:r.c]))
}

type wbuf struct {
	p []byte // buffer of data to write to
	c int    // current position in buffer of data
}

func newWriter(data []byte) *wbuf {
	return &wbuf{p: data, c: 0}
}

func (w *wbuf) Write(data []byte) (int, error) {
	n := copy(w.p[w.c:], data)
	if n < len(data) {
		return n, io.ErrShortWrite
	}
	w.c += n
	return n, nil
}

func (w *wbuf) writeByte(v byte) {
	w.p[w.c] = v
	w.c++
}

func (w *wbuf) writeI8(v int8) {
	w.p[w.c] = byte(v)
	w.c++
}

func (w *wbuf) writeI16(v int16) {
	beg := w.c
	w.c += 2
	binary.BigEndian.PutUint16(w.p[beg:w.c], uint16(v))
}

func (w *wbuf) writeI32(v int32) {
	beg := w.c
	w.c += 4
	binary.BigEndian.PutUint32(w.p[beg:w.c], uint32(v))
}

func (w *wbuf) writeI64(v int64) {
	beg := w.c
	w.c += 8
	binary.BigEndian.PutUint64(w.p[beg:w.c], uint64(v))
}

func (w *wbuf) writeF32(v float32) {
	beg := w.c
	w.c += 4
	binary.BigEndian.PutUint32(w.p[beg:w.c], math.Float32bits(v))
}

func (w *wbuf) writeF64(v float64) {
	beg := w.c
	w.c += 8
	binary.BigEndian.PutUint64(w.p[beg:w.c], math.Float64bits(v))
}
