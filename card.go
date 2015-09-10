// Copyright 2015 The astrogo Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fitsio

type Value interface{}

// Card is a record block in a Header
type Card struct {
	Name    string
	Value   Value
	Comment string
}
