// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

import (
	"bytes"
	"fmt"
	"io"
)

type Encoder interface {
	EncodeHDU(hdu HDU) error
}

// NewEncoder creates a new Encoder according to the capabilities of the underlying io.Writer
func NewEncoder(w io.Writer) Encoder {
	// FIXME(sbinet)
	// if ww, ok := w.(io.WriteSeeker); ok {
	// 	return &seekWriter{w: ww}
	// }
	return &streamEncoder{w: w}
}

// streamEncoder is a encoder which can not perform random access
// into the underlying Writer
type streamEncoder struct {
	w io.Writer
}

func (enc *streamEncoder) EncodeHDU(hdu HDU) error {
	var err error
	const (
		valind  = "= "
		nKEY    = 8
		nVALIND = 2
		nVAL    = 30
		nCOM    = 40
		nLINE   = 80
	)

	hdr := hdu.Header()
	nkeys := len(hdr.cards)
	buf := new(bytes.Buffer)

	buf.Grow(nkeys * nLINE)

	for i := range hdr.cards {
		card := &hdr.cards[i]
		bline, err := makeHeaderLine(card)
		if err != nil {
			return err
		}
		_, err = buf.Write(bline)
		if err != nil {
			return err
		}
	}

	{ // END
		bline, err := makeHeaderLine(&Card{Name: "END"})
		if err != nil {
			return err
		}
		_, err = buf.Write(bline)
		if err != nil {
			return err
		}
	}

	padsz := padBlock(buf.Len())
	if padsz > 0 {
		n, err := buf.Write(bytes.Repeat([]byte(" "), padsz))
		if err != nil {
			return fmt.Errorf("fitsio: error while padding header block: %v", err)
		}
		if n != padsz {
			return fmt.Errorf("fitsio: wrote %d bytes. expected %d. (padding)", n, padsz)
		}
	}

	alignsz := alignBlock(buf.Len())
	if alignsz != buf.Len() {
		return fmt.Errorf("fitsio: header not aligned (%d). expected %d.", buf.Len(), alignsz)
	}

	n, err := io.Copy(enc.w, buf)
	if err != nil {
		return fmt.Errorf("fitsio: error writing header block: %v", err)
	}
	if n != int64(alignsz) {
		return fmt.Errorf("fitsio: wrote %d bytes. expected %d", n, alignsz)
	}

	// write payload
	switch hdr.Type() {
	case IMAGE_HDU:
		img := hdu.(Image)
		err = enc.saveImage(img)
		if err != nil {
			return fmt.Errorf("fitsio: error encoding image: %v", err)
		}

	case BINARY_TBL:
		tbl := hdu.(*Table)
		err = enc.saveTable(tbl)
		if err != nil {
			return fmt.Errorf("fitsio: error encoding binary table: %v", err)
		}

	case ASCII_TBL:
		tbl := hdu.(*Table)
		err = enc.saveTable(tbl)
		if err != nil {
			return fmt.Errorf("fitsio: error encoding ascii table: %v", err)
		}

	case ANY_HDU:
		fallthrough
	default:
		return fmt.Errorf("fitsio: encoding for HDU [%v] not implemented", hdr.Type())
	}
	return err
}

func (enc *streamEncoder) saveImage(img Image) error {
	raw := img.Raw()
	n, err := enc.w.Write(raw)
	if err != nil {
		return err
	}
	if n != len(raw) {
		return fmt.Errorf("fitsio: wrote %d bytes. expected %d", n, len(raw))
	}

	padsz := padBlock(n)
	if padsz > 0 {
		n, err := enc.w.Write(make([]byte, padsz))
		if err != nil {
			return fmt.Errorf("fitsio: error while padding data-image block: %v", err)
		}
		if n != padsz {
			return fmt.Errorf("fitsio: wrote %d bytes. expected %d. (padding)", n, padsz)
		}
	}

	return err
}

func (enc *streamEncoder) saveTable(table *Table) error {
	ndata, err := enc.w.Write(table.data)
	if err != nil {
		return fmt.Errorf("fitsio: error writing table-data: %v", err)
	}
	if ndata != len(table.data) {
		return fmt.Errorf("fitsio: wrote %d bytes. expected %d", ndata, len(table.data))
	}

	nheap, err := enc.w.Write(table.heap)
	if err != nil {
		return fmt.Errorf("fitsio: error writing table-heap: %v", err)
	}
	if nheap != len(table.heap) {
		return fmt.Errorf("fitsio: wrote %d bytes. expected %d", nheap, len(table.heap))
	}

	// align to FITS block
	padsz := padBlock(ndata + nheap)
	if padsz > 0 {
		n := 0

		if table.binary {
			n, err = enc.w.Write(make([]byte, padsz))
		} else {
			n, err = enc.w.Write(bytes.Repeat([]byte(" "), padsz))
		}
		if err != nil {
			return fmt.Errorf("fitsio: error while padding table-data block: %v", err)
		}
		if n != padsz {
			return fmt.Errorf("fitsio: wrote %d bytes. expected %d. (padding)", n, padsz)
		}
	}

	return err
}
