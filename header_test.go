package fitsio

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestHeaderRW(t *testing.T) {
	curdir, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Chdir(curdir)

	workdir, err := ioutil.TempDir("", "go-fitsio-test-")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.RemoveAll(workdir)

	err = os.Chdir(workdir)
	if err != nil {
		t.Fatalf(err.Error())
	}

	table := struct {
		name    string
		version int
		cards   []Card
		bitpix  int
		axes    []int
		image   interface{}
	}{
		name:    "new.fits",
		version: 2,
		cards: []Card{
			{
				"EXTNAME",
				"primary hdu",
				"the primary HDU",
			},
			{
				"EXTVER",
				2,
				"the primary hdu version",
			},
			{
				"card_uint8",
				byte(42),
				"an uint8",
			},
			{
				"card_uint16",
				uint16(42),
				"an uint16",
			},
			{
				"card_uint32",
				uint32(42),
				"an uint32",
			},
			{
				"card_uint64",
				uint64(42),
				"an uint64",
			},
			{
				"card_int8",
				int8(42),
				"an int8",
			},
			{
				"card_int16",
				int16(42),
				"an int16",
			},
			{
				"card_int32",
				int32(42),
				"an int32",
			},
			{
				"card_int64",
				int64(42),
				"an int64",
			},
			{
				"card_int3264",
				int(42),
				"an int",
			},
			{
				"card_uintxx",
				uint(42),
				"an uint",
			},
			{
				"card_float32",
				float32(666),
				"a float32",
			},
			{
				"card_float64",
				float64(666),
				"a float64",
			},
			{
				"card_complex64",
				complex(float32(42), float32(66)),
				"a complex64",
			},
			{
				"card_complex128",
				complex(float64(42), float64(66)),
				"a complex128",
			},
		},
		bitpix: 8,
		axes:   []int{},
	}
	fname := "new.fits"
	for _, fct := range []func(){
		// create
		func() {
			w, err := os.Create(fname)
			if err != nil {
				t.Fatalf("error creating new file [%v]: %v", fname, err)
			}
			defer w.Close()

			f, err := Create(w)
			if err != nil {
				t.Fatalf("error creating new file [%v]: %v", fname, err)
			}
			defer f.Close()

			phdr := NewHeader(
				table.cards,
				IMAGE_HDU,
				table.bitpix,
				table.axes,
			)
			phdu, err := NewPrimaryHDU(phdr)
			if err != nil {
				t.Fatalf("error creating PHDU: %v", err)
			}
			defer phdu.Close()

			hdr := phdu.Header()
			if hdr.bitpix != table.bitpix {
				t.Fatalf("expected BITPIX=%v. got %v", table.bitpix, hdr.bitpix)
			}

			name := phdu.Name()
			if name != "primary hdu" {
				t.Fatalf("expected EXTNAME==%q. got %q", "primary hdu", name)
			}

			vers := phdu.Version()
			if vers != table.version {
				t.Fatalf("expected EXTVER==%v. got %v", table.version, vers)
			}

			card := hdr.Get("EXTNAME")
			if card == nil {
				t.Fatalf("error retrieving card [EXTNAME]")
			}
			if card.Comment != "the primary HDU" {
				t.Fatalf("expected EXTNAME.Comment==%q. got %q", "the primary HDU", card.Comment)
			}

			card = hdr.Get("EXTVER")
			if card == nil {
				t.Fatalf("error retrieving card [EXTVER]")
			}
			if card.Comment != "the primary hdu version" {
				t.Fatalf("expected EXTVER.Comment==%q. got %q", "the primary hdu version", card.Comment)

			}

			for _, ref := range table.cards {
				card := hdr.Get(ref.Name)
				if card == nil {
					t.Fatalf("error retrieving card [%v]", ref.Name)
				}
				rv := reflect.ValueOf(ref.Value)
				var val interface{}
				switch rv.Type().Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					val = int(rv.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					val = int(rv.Uint())
				case reflect.Float32, reflect.Float64:
					val = rv.Float()
				case reflect.Complex64, reflect.Complex128:
					val = rv.Complex()
				case reflect.String:
					val = ref.Value.(string)
				case reflect.Bool:
					val = ref.Value.(bool)
				}
				if !reflect.DeepEqual(card.Value, val) {
					t.Fatalf(
						"card %q. expected [%v](%T). got [%v](%T)",
						ref.Name,
						val, val,
						card.Value, card.Value,
					)
				}
				if card.Comment != ref.Comment {
					t.Fatalf("card %q. comment differ. expected %q. got %q", ref.Name, ref.Comment, card.Comment)
				}
			}

			card = hdr.Get("NOT THERE")
			if card != nil {
				t.Fatalf("expected no card. got [%v]", card)
			}

			err = f.Write(phdu)
			if err != nil {
				t.Fatalf("error writing hdu to file: %v", err)
			}
		},
		// read-back
		func() {
			r, err := os.Open(fname)
			if err != nil {
				t.Fatalf("error opening file [%v]: %v", fname, err)
			}
			defer r.Close()
			f, err := Open(r)
			if err != nil {
				buf, _ := ioutil.ReadFile(fname)
				t.Fatalf("error opening file [%v]: %v\nbuf=%s\n", fname, err, string(buf))
			}
			defer f.Close()

			hdu := f.HDU(0)
			hdr := hdu.Header()
			if hdr.bitpix != table.bitpix {
				t.Fatalf("expected BITPIX=%v. got %v", 8, hdr.bitpix)
			}

			name := hdu.Name()
			if name != "primary hdu" {
				t.Fatalf("expected EXTNAME==%q. got %q", "primary hdu", name)
			}

			vers := hdu.Version()
			if vers != table.version {
				t.Fatalf("expected EXTVER==%v. got %v", 2, vers)
			}

			card := hdr.Get("EXTNAME")
			if card == nil {
				t.Fatalf("error retrieving card [EXTNAME]")
			}
			if card.Comment != "the primary HDU" {
				t.Fatalf("expected EXTNAME.Comment==%q. got %q", "the primary HDU", card.Comment)
			}

			card = hdr.Get("EXTVER")
			if card == nil {
				t.Fatalf("error retrieving card [EXTVER]")
			}
			if card.Comment != "the primary hdu version" {
				t.Fatalf("expected EXTVER.Comment==%q. got %q", "the primary hdu version", card.Comment)

			}

			for _, ref := range table.cards {
				card := hdr.Get(ref.Name)
				if card == nil {
					t.Fatalf("error retrieving card [%v]", ref.Name)
				}

				rv := reflect.ValueOf(ref.Value)
				var val interface{}
				switch rv.Type().Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					val = int(rv.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					val = int(rv.Uint())
				case reflect.Float32, reflect.Float64:
					val = rv.Float()
				case reflect.Complex64, reflect.Complex128:
					val = rv.Complex()
				case reflect.String:
					val = ref.Value.(string)
				case reflect.Bool:
					val = ref.Value.(bool)
				}
				if !reflect.DeepEqual(card.Value, val) {
					t.Fatalf(
						"card %q. expected [%v](%T). got [%v](%T)",
						ref.Name,
						val, val,
						card.Value, card.Value,
					)
				}

				if card.Comment != ref.Comment {
					t.Fatalf("card %q. comment differ. expected %q. got %q", ref.Name, ref.Comment, card.Comment)
				}
			}

			card = hdr.Get("NOT THERE")
			if card != nil {
				t.Fatalf("expected no card. got [%v]", card)
			}
		},
	} {
		fct()
	}
}

func TestRWHeaderLine(t *testing.T) {
	for _, table := range []struct {
		line []byte
		card *Card
		err  error
	}{
		{
			line: []byte("SIMPLE  =                    T / file does conform to FITS standard             "),
			card: &Card{
				Name:    "SIMPLE",
				Value:   true,
				Comment: "file does conform to FITS standard",
			},
			err: nil,
		},
		{
			line: []byte("BITPIX  =                   16 / number of bits per data pixel                  "),
			card: &Card{
				Name:    "BITPIX",
				Value:   16,
				Comment: "number of bits per data pixel",
			},
			err: nil,
		},
		{
			line: []byte("EXTNAME = 'primary hdu'        / the primary HDU                                "),
			card: &Card{
				Name:    "EXTNAME",
				Value:   "primary hdu",
				Comment: "the primary HDU",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = 'a / '              / comment                                         "),
			card: &Card{
				Name:    "STRING",
				Value:   "a /",
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = ' a / '             / comment                                         "),
			card: &Card{
				Name:    "STRING",
				Value:   " a /",
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = ' a /              / comment                                        |'"),
			card: &Card{
				Name:    "STRING",
				Value:   " a /              / comment                                        |",
				Comment: "",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = ' a /              / comment                                         '"),
			card: &Card{
				Name:    "STRING",
				Value:   " a /              / comment",
				Comment: "",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = 'a / '''            / comment                                         "),
			card: &Card{
				Name:    "STRING",
				Value:   "a / '",
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("COMPLEX =        (42.0, 66.0) / comment                                         "),
			card: &Card{
				Name:    "COMPLEX",
				Value:   complex(42, 66),
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("COMPLEX =         (42.0,66.0) / comment                                         "),
			card: &Card{
				Name:    "COMPLEX",
				Value:   complex(42, 66),
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("COMPLEX =             (42,66) / comment                                         "),
			card: &Card{
				Name:    "COMPLEX",
				Value:   complex(42, 66),
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("COMPLEX =           (42.0,66) / comment                                         "),
			card: &Card{
				Name:    "COMPLEX",
				Value:   complex(42, 66),
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("COMPLEX =           (42,66.0) / comment                                         "),
			card: &Card{
				Name:    "COMPLEX",
				Value:   complex(42, 66),
				Comment: "comment",
			},
			err: nil,
		},
		{
			line: []byte("COMPLEX = (42.000,66.0000)    / comment                                         "),
			card: &Card{
				Name:    "COMPLEX",
				Value:   complex(42, 66),
				Comment: "comment",
			},
			err: nil,
		},
	} {
		card, err := parseHeaderLine(table.line)
		if !reflect.DeepEqual(err, table.err) {
			t.Fatalf("expected error [%v]. got: %v", table.err, err)
		}
		if !reflect.DeepEqual(card, table.card) {
			t.Fatalf("cards differ.\nexp= %#v\ngot= %#v", table.card, card)
		}

		line, err := makeHeaderLine(card)
		if err != nil {
			t.Fatalf("error making header-line: %v (%s)", err, string(line))
		}
	}

	for _, table := range []struct {
		line []byte
		err  error
	}{
		{
			line: []byte("SIMPLE= T / FITS file"),
			err:  fmt.Errorf("fitsio: invalid header line length"),
		},
		{
			line: []byte("STRING  = 'foo                   / comment                                      "),
			err:  fmt.Errorf(`fitsio: string ends prematurely ("'foo                   / comment                                      ")`),
		},
		{
			line: []byte("STRING  = 'foo ''                / comment                                      "),
			err:  fmt.Errorf(`fitsio: string ends prematurely ("'foo ''                / comment                                      ")`),
		},
	} {
		card, err := parseHeaderLine(table.line)
		if !reflect.DeepEqual(err, table.err) {
			t.Fatalf("expected error [%v]. got: %v\ncard=%#v", table.err, err, card)
		}
		if card != nil {
			t.Fatalf("expected a nil card. got= %#v", card)
		}
	}
}

func TestMakeHeaderLine(t *testing.T) {
	for _, table := range []struct {
		card *Card
		line []byte
		err  error
	}{
		{
			card: &Card{
				Name:    "SIMPLE",
				Value:   true,
				Comment: "file does conform to FITS standard",
			},
			line: []byte("SIMPLE  =                    T / file does conform to FITS standard             "),
			err:  nil,
		},
		{
			line: []byte("STRING  = ' a /              / no-comment                                    1&'CONTINUE  '2|      '                                                            COMMENT my-comment                                                              "),
			card: &Card{
				Name:    "STRING",
				Value:   " a /              / no-comment                                    12|",
				Comment: "my-comment",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = ' a /              / no-comment                                    1&'CONTINUE  '2|      '                                                            "),
			card: &Card{
				Name:    "STRING",
				Value:   " a /              / no-comment                                    12|",
				Comment: "",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = ' a /              / no-comment                                    |&'CONTINUE  '0123456789012345678901234567890123456789012345678901234567890123456&'CONTINUE  '7890123456789|'                                                      "),
			card: &Card{
				Name:    "STRING",
				Value:   " a /              / no-comment                                    |01234567890123456789012345678901234567890123456789012345678901234567890123456789|",
				Comment: "",
			},
			err: nil,
		},
		{
			line: []byte("STRING  = ' a /              / no-comment                                    |&'CONTINUE  '0123456789012345678901234567890123456789012345678901234567890123456&'CONTINUE  '7890123456789|'                                                      COMMENT my-comment                                                              "),
			card: &Card{
				Name:    "STRING",
				Value:   " a /              / no-comment                                    |01234567890123456789012345678901234567890123456789012345678901234567890123456789|",
				Comment: "my-comment",
			},
			err: nil,
		},
		{
			line: []byte("COMMENT *                                                                       "),
			card: &Card{
				Name:    "COMMENT",
				Value:   "",
				Comment: "*",
			},
			err: nil,
		},
	} {
		line, err := makeHeaderLine(table.card)
		if !reflect.DeepEqual(err, table.err) {
			t.Fatalf("expected error [%v]. got: %v\nline=%q", table.err, err, string(line))
		}
		if !reflect.DeepEqual(line, table.line) {
			t.Fatalf("bline differ.\nexp=%q\ngot=%q", string(table.line), string(line))
		}
	}
}

// EOF
