package fitsio

type Value interface{}

// Card is a record block in a Header
type Card struct {
	Name    string
	Value   Value
	Comment string
}
