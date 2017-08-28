// Copyright 2017 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"encoding/binary"
	"io"
	"math"
)

func readByte(r io.Reader, v *byte) error {
	var buf [1]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = buf[0]
	return nil
}

func readI8(r io.Reader, v *int8) error {
	var buf [1]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = int8(buf[0])
	return nil
}

func readI16(r io.Reader, v *int16) error {
	var buf [2]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = int16(binary.BigEndian.Uint16(buf[:]))
	return nil
}

func readI32(r io.Reader, v *int32) error {
	var buf [4]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = int32(binary.BigEndian.Uint32(buf[:]))
	return nil
}

func readI64(r io.Reader, v *int64) error {
	var buf [8]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = int64(binary.BigEndian.Uint64(buf[:]))
	return nil
}

func readF32(r io.Reader, v *float32) error {
	var buf [4]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = math.Float32frombits(binary.BigEndian.Uint32(buf[:]))
	return nil
}

func readF64(r io.Reader, v *float64) error {
	var buf [8]byte
	_, err := r.Read(buf[:])
	if err != nil {
		return err
	}
	*v = math.Float64frombits(binary.BigEndian.Uint64(buf[:]))
	return nil
}
