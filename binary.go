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
