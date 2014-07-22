package fitsio

import "io"

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}
