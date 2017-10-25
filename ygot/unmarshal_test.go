package ygot

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		inStruct   ValidatedGoStruct
		inFormat   JSONFormat
		inJSONPath string
		want       ValidatedGoStruct
		wantErr    string
	}{{
		name:     "simple schema JSON input",
		inStruct: &mapStructTestOne{},
		inFormat: Internal,
		want: &mapStructTestOne{
			Child: &mapStructTestOneChild{
				FieldOne: String("hello"),
				FieldTwo: Uint32(42),
			},
		},
		inJSONPath: filepath.Join(TestRoot, "testdata/emitjson_1.json-txt"),
	}, {
		name:     "schema with a list JSON output",
		inStruct: &mapStructTestFour{},
		inFormat: Internal,
		want: &mapStructTestFour{
			C: &mapStructTestFourC{
				ACLSet: map[string]*mapStructTestFourCACLSet{
					"n42": {Name: String("n42"), SecondValue: String("val")},
				},
			},
		},
		inJSONPath: filepath.Join(TestRoot, "testdata/emitjson_2.json-txt"),
	}, {
		name:     "simple schema IETF JSON output",
		inStruct: &mapStructTestOne{},
		inFormat: RFC7951,
		want: &mapStructTestOne{
			Child: &mapStructTestOneChild{
				FieldOne:  String("bar"),
				FieldTwo:  Uint32(84),
				FieldFive: Uint64(42),
			},
		},
		inJSONPath: filepath.Join(TestRoot, "testdata/emitjson1_ietf.json-txt"),
	}, {
		name:     "schema with list and enum IETF JSON",
		inStruct: &mapStructTestFour{},
		want: &mapStructTestFour{
			C: &mapStructTestFourC{
				ACLSet: map[string]*mapStructTestFourCACLSet{
					"n42": {Name: String("n42"), SecondValue: String("foo")},
				},
				OtherSet: map[ECTest]*mapStructTestFourCOtherSet{
					ECTestVALONE: {Name: ECTestVALONE},
					ECTestVALTWO: {Name: ECTestVALTWO},
				},
			},
		},
		inFormat:   RFC7951,
		inJSONPath: filepath.Join(TestRoot, "testdata/emitjson2_ietf.json-txt"),
	}}

	for _, tt := range tests {
		b, ioerr := ioutil.ReadFile(tt.inJSONPath)
		if ioerr != nil {
			t.Errorf("%s: ioutil.ReadFile(%s): could not open file: %v", tt.name, tt.inJSONPath, ioerr)
			continue
		}

		err := UnmarshalJSON(b, tt.inStruct, tt.inFormat)
		if errToString(err) != tt.wantErr {
			t.Errorf("%s: UnmarshalJSON(<json>, %v, %v): did not get expected error, got: %v, want: %v", tt.name, tt.inStruct, tt.inFormat, err, tt.wantErr)
			continue
		}

		if diff := pretty.Compare(tt.inStruct, tt.want); diff != "" {
			t.Errorf("%s: did not get expected output, diff(-got,+want):\n%s", tt.name, diff)
		}
	}
}
