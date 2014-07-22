package fitsio

import (
	"fmt"
)

type sectionWriter struct {
	buf []byte
	beg int
}

func (w *sectionWriter) Write(p []byte) (n int, err error) {
	n = copy(w.buf[w.beg:], p)
	if n < len(p) {
		return n, fmt.Errorf("fitsio: wrote %d bytes. expected %d", n, len(p))
	}
	w.beg += n
	return
}
