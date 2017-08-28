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

func (r *rbuf) readBool(v *bool) {
	b := r.p[r.c]
	r.c++
	if b == 0 {
		*v = false
	} else {
		*v = true
	}
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

func (r *rbuf) readInt(v *int) {
	beg := r.c
	r.c += 8
	*v = int(int64(binary.BigEndian.Uint64(r.p[beg:r.c])))
}

func (r *rbuf) readU8(v *uint8) {
	*v = r.p[r.c]
	r.c++
}

func (r *rbuf) readU16(v *uint16) {
	beg := r.c
	r.c += 2
	*v = binary.BigEndian.Uint16(r.p[beg:r.c])
}

func (r *rbuf) readU32(v *uint32) {
	beg := r.c
	r.c += 4
	*v = binary.BigEndian.Uint32(r.p[beg:r.c])
}

func (r *rbuf) readU64(v *uint64) {
	beg := r.c
	r.c += 8
	*v = binary.BigEndian.Uint64(r.p[beg:r.c])
}

func (r *rbuf) readUint(v *uint) {
	beg := r.c
	r.c += 8
	*v = uint(binary.BigEndian.Uint64(r.p[beg:r.c]))
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

func (r *rbuf) readC64(v *complex64) {
	beg := r.c
	r.c += 4
	vr := math.Float32frombits(binary.BigEndian.Uint32(r.p[beg:r.c]))

	beg = r.c
	r.c += 4
	vi := math.Float32frombits(binary.BigEndian.Uint32(r.p[beg:r.c]))

	*v = complex(vr, vi)
}

func (r *rbuf) readC128(v *complex128) {
	beg := r.c
	r.c += 8
	vr := math.Float64frombits(binary.BigEndian.Uint64(r.p[beg:r.c]))

	beg = r.c
	r.c += 8
	vi := math.Float64frombits(binary.BigEndian.Uint64(r.p[beg:r.c]))

	*v = complex(vr, vi)
}

func (r *rbuf) readBools(vs []bool) {
	for i := range vs {
		r.readBool(&vs[i])
	}
}

func (r *rbuf) readI8s(vs []int8) {
	for i := range vs {
		r.readI8(&vs[i])
	}
}

func (r *rbuf) readI16s(vs []int16) {
	for i := range vs {
		r.readI16(&vs[i])
	}
}

func (r *rbuf) readI32s(vs []int32) {
	for i := range vs {
		r.readI32(&vs[i])
	}
}

func (r *rbuf) readI64s(vs []int64) {
	for i := range vs {
		r.readI64(&vs[i])
	}
}

func (r *rbuf) readInts(vs []int) {
	for i := range vs {
		r.readInt(&vs[i])
	}
}

func (r *rbuf) readU8s(vs []uint8) {
	for i := range vs {
		r.readU8(&vs[i])
	}
}

func (r *rbuf) readU16s(vs []uint16) {
	for i := range vs {
		r.readU16(&vs[i])
	}
}

func (r *rbuf) readU32s(vs []uint32) {
	for i := range vs {
		r.readU32(&vs[i])
	}
}

func (r *rbuf) readU64s(vs []uint64) {
	for i := range vs {
		r.readU64(&vs[i])
	}
}

func (r *rbuf) readUints(vs []uint) {
	for i := range vs {
		r.readUint(&vs[i])
	}
}

func (r *rbuf) readF32s(vs []float32) {
	for i := range vs {
		r.readF32(&vs[i])
	}
}

func (r *rbuf) readF64s(vs []float64) {
	for i := range vs {
		r.readF64(&vs[i])
	}
}

func (r *rbuf) readC64s(vs []complex64) {
	for i := range vs {
		r.readC64(&vs[i])
	}
}

func (r *rbuf) readC128s(vs []complex128) {
	for i := range vs {
		r.readC128(&vs[i])
	}
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

func (w *wbuf) bytes() []byte {
	return w.p
}

func (w *wbuf) writeByte(v byte) {
	w.p[w.c] = v
	w.c++
}

func (w *wbuf) writeBool(v bool) {
	var b byte
	if v {
		b = 1
	}
	w.p[w.c] = b
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

func (w *wbuf) writeInt(v int) {
	beg := w.c
	w.c += 8
	binary.BigEndian.PutUint64(w.p[beg:w.c], uint64(v))
}

func (w *wbuf) writeU8(v uint8) {
	w.p[w.c] = v
	w.c++
}

func (w *wbuf) writeU16(v uint16) {
	beg := w.c
	w.c += 2
	binary.BigEndian.PutUint16(w.p[beg:w.c], v)
}

func (w *wbuf) writeU32(v uint32) {
	beg := w.c
	w.c += 4
	binary.BigEndian.PutUint32(w.p[beg:w.c], v)
}

func (w *wbuf) writeU64(v uint64) {
	beg := w.c
	w.c += 8
	binary.BigEndian.PutUint64(w.p[beg:w.c], v)
}

func (w *wbuf) writeUint(v uint) {
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

func (w *wbuf) writeC64(v complex64) {
	beg := w.c
	w.c += 4
	binary.BigEndian.PutUint32(w.p[beg:w.c], math.Float32bits(real(v)))
	beg = w.c
	w.c += 4
	binary.BigEndian.PutUint32(w.p[beg:w.c], math.Float32bits(imag(v)))
}

func (w *wbuf) writeC128(v complex128) {
	beg := w.c
	w.c += 8
	binary.BigEndian.PutUint64(w.p[beg:w.c], math.Float64bits(real(v)))
	beg = w.c
	w.c += 8
	binary.BigEndian.PutUint64(w.p[beg:w.c], math.Float64bits(imag(v)))
}

func (w *wbuf) writeBools(vs []bool) {
	for _, v := range vs {
		w.writeBool(v)
	}
}

func (w *wbuf) writeI8s(vs []int8) {
	for _, v := range vs {
		w.writeI8(v)
	}
}

func (w *wbuf) writeI16s(vs []int16) {
	for _, v := range vs {
		w.writeI16(v)
	}
}

func (w *wbuf) writeI32s(vs []int32) {
	for _, v := range vs {
		w.writeI32(v)
	}
}

func (w *wbuf) writeI64s(vs []int64) {
	for _, v := range vs {
		w.writeI64(v)
	}
}

func (w *wbuf) writeU8s(vs []uint8) {
	for _, v := range vs {
		w.writeU8(v)
	}
}

func (w *wbuf) writeU16s(vs []uint16) {
	for _, v := range vs {
		w.writeU16(v)
	}
}

func (w *wbuf) writeU32s(vs []uint32) {
	for _, v := range vs {
		w.writeU32(v)
	}
}

func (w *wbuf) writeU64s(vs []uint64) {
	for _, v := range vs {
		w.writeU64(v)
	}
}

func (w *wbuf) writeF32s(vs []float32) {
	for _, v := range vs {
		w.writeF32(v)
	}
}

func (w *wbuf) writeF64s(vs []float64) {
	for _, v := range vs {
		w.writeF64(v)
	}
}

func (w *wbuf) writeC64s(vs []complex64) {
	for _, v := range vs {
		w.writeC64(v)
	}
}

func (w *wbuf) writeC128s(vs []complex128) {
	for _, v := range vs {
		w.writeC128(v)
	}
}

func (w *wbuf) writeUints(vs []uint) {
	for _, v := range vs {
		w.writeUint(v)
	}
}

func (w *wbuf) writeInts(vs []int) {
	for _, v := range vs {
		w.writeInt(v)
	}
}
