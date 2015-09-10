// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
