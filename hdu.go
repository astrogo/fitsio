package fitsio

import (
	"fmt"
)

// HDUType is the type of a Header-Data Unit
type HDUType int

const (
	IMAGE_HDU  HDUType = iota // Primary Array or IMAGE HDU
	ASCII_TBL                 // ASCII table HDU
	BINARY_TBL                // Binary table HDU
	ANY_HDU                   // matches any HDU type
)

func (htype HDUType) String() string {
	switch htype {
	case IMAGE_HDU:
		return "IMAGE"
	case ASCII_TBL:
		return "TABLE"
	case BINARY_TBL:
		return "BINTABLE"
	case ANY_HDU:
		return "ANY_HDU"
	default:
		panic(fmt.Errorf("invalid HDU Type value (%v)", int(htype)))
	}
}

// HDU is a "Header-Data Unit" block
type HDU interface {
	Close() error
	Type() HDUType
	Name() string
	Version() int
	Header() *Header
}
