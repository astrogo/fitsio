//+build ignore

package main

import "fmt"

type Descriptor struct {
	Type  string
	Len   []int
	Range [2]int
}

func main() {

	const (
		min = 2
		max = 50
	)

	descr := []Descriptor{
		{
			Type:  "byte",
			Len:   []int{2, 1024},
			Range: [2]int{min, max},
		},
		{
			Type:  "bool",
			Range: [2]int{min, max},
		},
		// {
		// 	Type:  "int8",
		// 	Range: [2]int{min, max},
		// },
		{
			Type:  "int16",
			Range: [2]int{min, max},
		},
		{
			Type:  "int32",
			Range: [2]int{min, max},
		},
		{
			Type:  "int64",
			Range: [2]int{min, max},
		},
		{
			Type:  "uint8",
			Range: [2]int{min, max},
		},
		// {
		// 	Type:  "uint16",
		// 	Range: [2]int{min, max},
		// },
		// {
		// 	Type:  "uint32",
		// 	Range: [2]int{min, max},
		// },
		// {
		// 	Type:  "uint64",
		// 	Range: [2]int{min, max},
		// },
		{
			Type:  "float32",
			Len:   []int{2, 3, 4, 5, 376},
			Range: [2]int{min, max},
		},
		{
			Type:  "float64",
			Range: [2]int{min, max},
		},
		{
			Type:  "complex64",
			Range: [2]int{min, max},
		},
		{
			Type:  "complex128",
			Range: [2]int{min, max},
		},
	}

	fmt.Printf(`// generated with "go run ./gen-arraytypes.go"

package fitsio

import (
	"reflect"
)

// this is a hack, waiting for reflect.ArrayOf to appear in a go release.
var g_arrayTypes map[string]reflect.Type

func init() {
	g_arrayTypes = map[string]reflect.Type{
`)

	set := make(map[string]struct{})
	for _, d := range descr {
		//fmt.Printf("\t\ttyp_%s: {\n", string(d.Name))
		template := "\t\t%[1]q: reflect.TypeOf((*%[1]s)(nil)).Elem(),\n"
		for i := d.Range[0]; i <= d.Range[1]; i++ {
			tname := fmt.Sprintf("[%d]%s", i, d.Type)
			if _, dup := set[tname]; dup {
				continue
			}
			fmt.Printf(template, tname)
			set[tname] = struct{}{}
		}
		if len(d.Len) > 0 {
			for _, i := range d.Len {
				if d.Range[0] <= i && i <= d.Range[1] {
					continue
				}
				tname := fmt.Sprintf("[%d]%s", i, d.Type)
				if _, dup := set[tname]; dup {
					continue
				}
				fmt.Printf(template, tname)
				set[tname] = struct{}{}
			}
		}
	}
	fmt.Printf("\t}\n}\n")
}
