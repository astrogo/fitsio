package fitsio

type any interface{}

var g_tables = []struct {
	fname string
	hdus  []HDU
	tuple [][][]any
	maps  []map[string]any
	types []interface{}
}{
	{
		fname: "testdata/swp06542llg.fits",
		hdus: []HDU{
			&primaryHDU{imageHDU{
				hdr: *newHeader(
					[]Card{
						{
							Name:    "SIMPLE",
							Value:   true,
							Comment: "Standard FITS format",
						},
						{
							Name:    "BITPIX",
							Value:   8,
							Comment: "",
						},
						{
							Name:    "NAXIS",
							Value:   0,
							Comment: "no data in main file",
						},
						{
							Name:    "EXTEND",
							Value:   true,
							Comment: "Extensions may exist",
						},
						{
							Name:    "FILENAME",
							Value:   "swp06542llg",
							Comment: "original name of input file",
						},
						{
							Name:    "TELESCOP",
							Value:   "IUE",
							Comment: "International Ultraviolet Explorer",
						},
						{
							Name:    "ORIGIN",
							Value:   "GODDARD",
							Comment: "Tape writing location",
						},
						{
							Name:    "CAMERA",
							Value:   3,
							Comment: "IUE camera number",
						},
						{
							Name:    "IMAGE",
							Value:   6542,
							Comment: "IUE image sequence number",
						},
						{
							Name:    "APERTURE",
							Value:   "",
							Comment: "Aperture",
						},
						{
							Name:    "DISPERSN",
							Value:   "LOW",
							Comment: "IUE spectrograph dispersion",
						},
						{
							Name:    "DATE-OBS",
							Value:   "nn/nn/nn",
							Comment: "Observation date (dd/mm/yy)",
						},
						{
							Name:    "DATE-PRO",
							Value:   "nn/nn/nn",
							Comment: "Processing date (dd/mm/yy)",
						},
						{
							Name:    "DATE",
							Value:   "18-Feb-1993",
							Comment: "Date file was written (dd/mm/yy)",
						},
						{
							Name:    "RA",
							Value:   0.0,
							Comment: "Right Ascension in degrees",
						},
						{
							Name:    "DEC",
							Value:   0.0,
							Comment: "Declination in degrees",
						},
						{
							Name:    "EQUINOX",
							Value:   1950.0,
							Comment: "Epoch for coordinates (years)",
						},
						{
							Name:    "THDA-RES",
							Value:   0.0,
							Comment: "THDA at time of read",
						},
						{
							Name:    "THDA-SPE",
							Value:   0.0,
							Comment: "THDA at end of exposure",
						},
					},
					IMAGE_HDU,
					8,
					[]int{},
				),
			}},
			&Table{
				hdr: *newHeader(
					[]Card{
						{
							Name:    "XTENSION",
							Value:   "BINTABLE",
							Comment: "Extension type",
						},
						{
							Name:    "BITPIX",
							Value:   8,
							Comment: "binary data",
						},
						{
							Name:    "NAXIS",
							Value:   2,
							Comment: "Number of Axes",
						},
						{
							Name:    "NAXIS1",
							Value:   7532,
							Comment: "width of table in bytes",
						},
						{
							Name:    "NAXIS2",
							Value:   1,
							Comment: "Number of entries in table",
						},
						{
							Name:    "PCOUNT",
							Value:   0,
							Comment: "Number of parameters/group",
						},
						{
							Name:    "GCOUNT",
							Value:   1,
							Comment: "Number of groups",
						},
						{
							Name:    "TFIELDS",
							Value:   9,
							Comment: "Number of fields in each row",
						},
						{
							Name:    "EXTNAME",
							Value:   "IUE MELO",
							Comment: "name of table (?)",
						},
						{
							Name:    "TFORM1",
							Value:   "1I",
							Comment: "Count and data type of field 1",
						},
						{
							Name:    "TTYPE1",
							Value:   "ORDER",
							Comment: "spectral order (low dispersion = 1)",
						},
						{
							Name:    "TUNIT1",
							Value:   "",
							Comment: "unitless",
						},
						{
							Name:    "TFORM2",
							Value:   "1I",
							Comment: "field 2 has one 2-byte integer",
						},
						{
							Name:    "TTYPE2",
							Value:   "NPTS",
							Comment: "number of non-zero points in each vector",
						},
						{
							Name:    "TUNIT2",
							Value:   "",
							Comment: "unitless",
						},
						{
							Name:    "TFORM3",
							Value:   "1E",
							Comment: "Count and data type of field 3",
						},
						{
							Name:    "TTYPE3",
							Value:   "LAMBDA",
							Comment: "3rd field is starting wavelength",
						},
						{
							Name:    "TUNIT3",
							Value:   "ANGSTROM",
							Comment: "unit is angstrom",
						},
						{
							Name:    "TFORM4",
							Value:   "1E",
							Comment: "Count and Type of 4th field",
						},
						{
							Name:    "TTYPE4",
							Value:   "DELTAW",
							Comment: "4th field is wavelength increment",
						},
						{
							Name:    "TUNIT4",
							Value:   "ANGSTROM",
							Comment: "unit is angstrom",
						},
						{
							Name:    "TFORM5",
							Value:   "376E",
							Comment: "Count and Type of 5th field",
						},
						{
							Name:    "TTYPE5",
							Value:   "GROSS",
							Comment: "5th field is gross flux array",
						},
						{
							Name:    "TUNIT5",
							Value:   "FN",
							Comment: "unit is IUE FN",
						},
						{
							Name:    "TFORM6",
							Value:   "376E",
							Comment: "Count and Type of 6th field",
						},
						{
							Name:    "TTYPE6",
							Value:   "BACK",
							Comment: "6th field is background flux array",
						},
						{
							Name:    "TUNIT6",
							Value:   "FN",
							Comment: "unit is IUE FN",
						},
						{
							Name:    "TFORM7",
							Value:   "376E",
							Comment: "Count and Type of 7th field",
						},
						{
							Name:    "TTYPE7",
							Value:   "NET",
							Comment: "7th field is net flux array",
						},
						{
							Name:    "TUNIT7",
							Value:   "ERGS",
							Comment: "unit is IUE FN",
						},
						{
							Name:    "TFORM8",
							Value:   "376E",
							Comment: "Count and Type of 8th field",
						},
						{
							Name:    "TTYPE8",
							Value:   "ABNET",
							Comment: "absolutely calibrated net flux array",
						},
						{
							Name:    "TUNIT8",
							Value:   "ERGS",
							Comment: "unit is ergs/cm2/sec/angstrom",
						},
						{
							Name:    "TFORM9",
							Value:   "376E",
							Comment: "Count and Type of 9th field",
						},
						{
							Name:    "TTYPE9",
							Value:   "EPSILONS",
							Comment: "9th field is epsilons",
						},
						{
							Name:    "TUNIT9",
							Value:   "",
							Comment: "unitless",
						},
					},
					BINARY_TBL,
					8,
					[]int{},
				),
			},
		},
		tuple: [][][]any{
			nil,
			{
				// row-0
				{
					int16(1),
					int16(376), float32(1000.8), float32(2.6515958), g_data_gross,
					g_data_back, g_data_net, g_data_abnet,
					g_data_epsilons,
				},
			},
		},
		maps: []map[string]any{
			{},
			{
				"ORDER":    int16(1),
				"NPTS":     int16(376),
				"LAMBDA":   float32(1000.8),
				"DELTAW":   float32(2.6515958),
				"GROSS":    g_data_gross,
				"BACK":     g_data_back,
				"NET":      g_data_net,
				"ABNET":    g_data_abnet,
				"EPSILONS": g_data_epsilons,
			},
		},
		types: []interface{}{
			nil,
			struct {
				Order    int16        `fits:"ORDER"`
				Npts     int16        `fits:"NPTS"`
				DeltaW   float32      `fits:"DELTAW"` // switch order of deltaw w/ lambda
				Lambda   float32      `fits:"LAMBDA"`
				Gross    [376]float32 `fits:"GROSS"`
				Back     [376]float32 `fits:"BACK"`
				Net      [376]float32 `fits:"NET"`
				ABNET    [376]float32 // test w/o struct-tag
				EPSILONS [376]float32 // ditto
			}{},
		},
	},
	{
		fname: "testdata/file001.fits",
		hdus: []HDU{
			&primaryHDU{imageHDU{
				hdr: *newHeader(
					[]Card{
						{
							Name:    "SIMPLE",
							Value:   true,
							Comment: "STANDARD FITS FORMAT (REV OCT 1981)",
						},
						{
							Name:    "BITPIX",
							Value:   8,
							Comment: "CHARACTER INFORMATION",
						},
						{
							Name:    "NAXIS",
							Value:   0,
							Comment: "NO IMAGE DATA ARRAY PRESENT",
						},
						{
							Name:    "EXTEND",
							Value:   true,
							Comment: "THERE IS AN EXTENSION",
						},
						{
							Name:    "ORIGIN",
							Value:   "ESO",
							Comment: "EUROPEAN SOUTHERN OBSERVATORY",
						},
						{
							Name:    "OBJECT",
							Value:   "SNG - CAT.",
							Comment: "THE IDENTIFIER",
						},
						{
							Name:    "DATE",
							Value:   "27/ 5/84",
							Comment: "DATE THIS TAPE WRITTEN DD/MM/YY",
						},
					},
					IMAGE_HDU,
					8,
					[]int{},
				),
			}},
			&Table{
				hdr: *newHeader(
					[]Card{
						{
							Name:    "XTENSION",
							Value:   "TABLE",
							Comment: "TABLE EXTENSION",
						},
						{
							Name:    "BITPIX",
							Value:   8,
							Comment: "CHARACTER INFORMATION",
						},
						{
							Name:    "NAXIS",
							Value:   2,
							Comment: "SIMPLE 2-D MATRIX",
						},
						{
							Name:    "NAXIS1",
							Value:   98,
							Comment: "NO. OF CHARACTERS PER ROW",
						},
						{
							Name:    "NAXIS2",
							Value:   10,
							Comment: "NO. OF ROWS",
						},
						{
							Name:    "PCOUNT",
							Value:   0,
							Comment: "RANDOM PARAMETER COUNT",
						},
						{
							Name:    "GCOUNT",
							Value:   1,
							Comment: "GROUP COUNT",
						},
						{
							Name:    "TFIELDS",
							Value:   7,
							Comment: "NO. OF FIELDS PER ROW",
						},
						{
							Name:    "TTYPE1",
							Value:   "IDEN.",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL1",
							Value:   1,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM1",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL1",
							Value:   "",
							Comment: "NULL VALUE",
						},
						{
							Name:    "TTYPE2",
							Value:   "RA",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL2",
							Value:   15,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM2",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL2",
							Value:   "",
							Comment: "NULL VALUE",
						},
						{
							Name:    "TTYPE3",
							Value:   "DEC",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL3",
							Value:   29,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM3",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL3",
							Value:   "",
							Comment: "NULL VALUE",
						},
						{
							Name:    "TTYPE4",
							Value:   "TYPE",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL4",
							Value:   43,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM4",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL4",
							Value:   "",
							Comment: "NULL VALUE",
						},
						{
							Name:    "TTYPE5",
							Value:   "D25",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL5",
							Value:   57,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM5",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL5",
							Value:   "",
							Comment: "NULL VALUE",
						},
						{
							Name:    "TTYPE6",
							Value:   "INCL.",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL6",
							Value:   71,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM6",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL6",
							Value:   "",
							Comment: "NULL VALUE",
						},
						{
							Name:    "TTYPE7",
							Value:   "RV",
							Comment: "NAME OF ROW",
						},
						{
							Name:    "TBCOL7",
							Value:   85,
							Comment: "BEGINNING COLUMN OF THE FIELD",
						},
						{
							Name:    "TFORM7",
							Value:   "E14.7",
							Comment: "FORMAT",
						},
						{
							Name:    "TNULL7",
							Value:   "",
							Comment: "NULL VALUE",
						},
					},
					ASCII_TBL,
					8,
					[]int{98, 10},
				),
			},
		},
		tuple: [][][]any{
			nil,
			{
				{
					-.1116590E+04, .1128000E+02, .5956670E+02, .3000000E+01,
					.7789999E+02, .1200000E+02, .0000000E+00,
				},
				{
					-.1109540E+04, .1115667E+02, .5430000E+02, .3000000E+01,
					.4000000E+02, .1200000E+02, .0000000E+00,
				},
				{
					-.3402850E+03, .3668300E+01, -.2801670E+02, .3000000E+01,
					.7000000E+02, .3000000E+02, .4060000E+04,
				},
				{
					.5360000E+03, .1393330E+01, .3445000E+02, .3500000E+01,
					.2229000E+03, .5990000E+02, .5160000E+04,
				},
				{
					.3177000E+04, .1023000E+02, .2136670E+02, .3000000E+01,
					.9960001E+02, .3600000E+02, .1220000E+04,
				},
				{
					.3627000E+04, .1129333E+02, .1326670E+02, .3200000E+01,
					.5226000E+03, .5990000E+02, .6970000E+03,
				},
				{
					.3756000E+04, .1156667E+02, .5456670E+02, .4200000E+01,
					.2619000E+03, .5670000E+02, .1071000E+04,
				},
				{
					.5457000E+04, .1402500E+02, .5458300E+02, .6200000E+01,
					.1614900E+04, .1220000E+02, .2660000E+03,
				},
				{
					.7292000E+04, .2243500E+02, .3005000E+02, .9500000E+01,
					.1283000E+03, .3560000E+02, .9340000E+03,
				},
				{
					.1423700E+05, .1336330E+02, -.2086700E+02, .3500000E+01,
					.1170000E+03, .4220000E+02, .0000000E+00,
				},
			},
		},
		maps: []map[string]any{
			{},
			{
				"IDEN.": -1116.59,
				"RA":    11.28,
				"DEC":   59.5667,
				"TYPE":  float64(3),
				"D25":   77.89999,
				"INCL.": float64(12),
				"RV":    float64(0),
				//"NOT-THERE": 0.0,
			},
		},
		types: []interface{}{
			nil,
			struct {
				Iden float64 `fits:"IDEN."`
				Ra   float64 `fits:"RA"`
				Dec  float64 `fits:"DEC"`
				Type float64 `fits:"TYPE"`
				D25  float64 `fits:"D25"`
				Incl float64 `fits:"INCL."`
				RV   float64 // test w/o struct-tag
				//X_NotThere float64 `fits:"NOT_THERE"`
			}{},
		},
	},
}
