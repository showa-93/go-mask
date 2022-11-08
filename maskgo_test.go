package maskgo

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/goccy/go-reflect"

	"github.com/ggwhite/go-masker"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	type stringTest struct {
		Usagi string
	}
	type stringPtrTest struct {
		Usagi *string
	}
	type stringSliceTest struct {
		Usagi []string
	}
	type stringSlicePtrTest struct {
		Usagi *[]string
	}
	type intTest struct {
		Usagi int
	}
	type intPtrTest struct {
		Usagi *int
	}
	type intSliceTest struct {
		Usagi []int
	}
	type intSlicePtrTest struct {
		Usagi *[]int
	}
	type float64Test struct {
		Usagi float64
	}
	type float64PtrTest struct {
		Usagi *float64
	}
	type float64SliceTest struct {
		Usagi []float64
	}
	type float64SlicePtrTest struct {
		Usagi *[]float64
	}
	type boolTest struct {
		Usagi bool
	}
	type boolPtrTest struct {
		Usagi *bool
	}
	type mapStringToStringTest struct {
		Usagi map[string]string
	}
	type mapStringToIntTest struct {
		Usagi map[string]int
	}
	type mapStringToFloat64Test struct {
		Usagi map[string]float64
	}
	type mapIntToStringTest struct {
		Usagi map[int]string
	}
	type mapIntToIntTest struct {
		Usagi map[int]int
	}
	type mapIntToFloat64Test struct {
		Usagi map[int]float64
	}
	type mapStructToStringTest struct {
		Usagi map[stringTest]string
	}
	type mapStructToIntTest struct {
		Usagi map[stringTest]int
	}
	type mapStructToFloat64Test struct {
		Usagi map[stringTest]float64
	}
	type structTest struct {
		StringTest      stringTest
		StringSliceTest stringSliceTest
	}
	type structSliceTest struct {
		SliceTest []stringTest
	}
	type unexportedTest struct {
		usagi string
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string": {
			input: "ヤハッ！",
			want:  "ヤハッ！",
		},
		"string empty": {
			input: "",
			want:  "",
		},
		"string ptr": {
			input: convertStringPtr("ヤハッ！"),
			want:  convertStringPtr("ヤハッ！"),
		},
		"nil string ptr": {
			input: (*string)(nil),
			want:  (*string)(nil),
		},
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！"},
			want:  &stringTest{Usagi: "ヤハッ！"},
		},
		"string empty fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
		},
		"string slice": {
			input: []string{"ハァ？", "ウラ", "フゥン"},
			want:  []string{"ハァ？", "ウラ", "フゥン"},
		},
		"nil string slice": {
			input: ([]string)(nil),
			want:  ([]string)(nil),
		},
		"string slice ptr": {
			input: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"}),
			want:  convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"}),
		},
		"nil string slice ptr": {
			input: (*[]string)(nil),
			want:  (*[]string)(nil),
		},
		"map string to string": {
			input: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"},
			want:  map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"},
		},
		"nil map string to string": {
			input: (map[string]string)(nil),
			want:  (map[string]string)(nil),
		},
		"string ptr fields": {
			input: &stringPtrTest{Usagi: convertStringPtr("ヤハッ！")},
			want:  &stringPtrTest{Usagi: convertStringPtr("ヤハッ！")},
		},
		"nil string ptr fields": {
			input: &stringPtrTest{},
			want:  &stringPtrTest{Usagi: nil},
		},
		"string slice fields": {
			input: &stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
			want:  &stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"})},
			want:  &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"})},
		},
		"nil string slice ptr fields": {
			input: &stringSlicePtrTest{},
			want:  &stringSlicePtrTest{Usagi: (*[]string)(nil)},
		},
		"int": {
			input: 20190122,
			want:  20190122,
		},
		"zero int": {
			input: 0,
			want:  0,
		},
		"int ptr": {
			input: convertIntPtr(20190122),
			want:  convertIntPtr(20190122),
		},
		"nil int ptr": {
			input: (*int)(nil),
			want:  (*int)(nil),
		},
		"int slice": {
			input: []int{20190122, 20200501, 20200501},
			want:  []int{20190122, 20200501, 20200501},
		},
		"nil int slice": {
			input: ([]int)(nil),
			want:  ([]int)(nil),
		},
		"int slice ptr": {
			input: convertIntSlicePtr([]int{20190122, 20200501, 20200501}),
			want:  convertIntSlicePtr([]int{20190122, 20200501, 20200501}),
		},
		"nil int slice ptr": {
			input: (*[]int)(nil),
			want:  (*[]int)(nil),
		},
		"int fields": {
			input: &intTest{Usagi: 20190122},
			want:  &intTest{Usagi: 20190122},
		},
		"zero int fields": {
			input: &intTest{},
			want:  &intTest{Usagi: 0},
		},
		"int ptr fields": {
			input: &intPtrTest{Usagi: convertIntPtr(20190122)},
			want:  &intPtrTest{Usagi: convertIntPtr(20190122)},
		},
		"nil int ptr fields": {
			input: &intPtrTest{},
			want:  &intPtrTest{Usagi: nil},
		},
		"int slice fields": {
			input: &intSliceTest{Usagi: []int{20190122, 20200501, 20200501}},
			want:  &intSliceTest{Usagi: []int{20190122, 20200501, 20200501}},
		},
		"nil int slice fields": {
			input: &intSliceTest{},
			want:  &intSliceTest{Usagi: ([]int)(nil)},
		},
		"int slice ptr fields": {
			input: &intSlicePtrTest{Usagi: convertIntSlicePtr([]int{20190122, 20200501, 20200501})},
			want:  &intSlicePtrTest{Usagi: convertIntSlicePtr([]int{20190122, 20200501, 20200501})},
		},
		"nil int slice ptr fields": {
			input: &intSlicePtrTest{},
			want:  &intSlicePtrTest{Usagi: (*[]int)(nil)},
		},
		"float64": {
			input: 20190122,
			want:  20190122,
		},
		"zero float64": {
			input: 0,
			want:  0,
		},
		"float64 ptr": {
			input: convertFloat64Ptr(20190122),
			want:  convertFloat64Ptr(20190122),
		},
		"nil float64 ptr": {
			input: (*float64)(nil),
			want:  (*float64)(nil),
		},
		"float64 slice": {
			input: []float64{20190122, 20200501, 20200501},
			want:  []float64{20190122, 20200501, 20200501},
		},
		"nil float64 slice": {
			input: ([]float64)(nil),
			want:  ([]float64)(nil),
		},
		"float64 slice ptr": {
			input: convertFloat64SlicePtr([]float64{20190122, 20200501, 20200501}),
			want:  convertFloat64SlicePtr([]float64{20190122, 20200501, 20200501}),
		},
		"nil float64 slice ptr": {
			input: (*[]float64)(nil),
			want:  (*[]float64)(nil),
		},
		"float64 fields": {
			input: &float64Test{Usagi: 20190122},
			want:  &float64Test{Usagi: 20190122},
		},
		"zero float64 fields": {
			input: &float64Test{},
			want:  &float64Test{Usagi: 0},
		},
		"float64 ptr fields": {
			input: &float64PtrTest{Usagi: convertFloat64Ptr(20190122)},
			want:  &float64PtrTest{Usagi: convertFloat64Ptr(20190122)},
		},
		"nil float64 ptr fields": {
			input: &float64PtrTest{},
			want:  &float64PtrTest{Usagi: nil},
		},
		"float64 slice fields": {
			input: &float64SliceTest{Usagi: []float64{20190122, 20200501, 20200501}},
			want:  &float64SliceTest{Usagi: []float64{20190122, 20200501, 20200501}},
		},
		"nil float64 slice fields": {
			input: &float64SliceTest{},
			want:  &float64SliceTest{Usagi: ([]float64)(nil)},
		},
		"float64 slice ptr fields": {
			input: &float64SlicePtrTest{Usagi: convertFloat64SlicePtr([]float64{20190122, 20200501, 20200501})},
			want:  &float64SlicePtrTest{Usagi: convertFloat64SlicePtr([]float64{20190122, 20200501, 20200501})},
		},
		"nil float64 slice ptr fields": {
			input: &float64SlicePtrTest{},
			want:  &float64SlicePtrTest{Usagi: (*[]float64)(nil)},
		},
		"bool fields": {
			input: &boolTest{Usagi: true},
			want:  &boolTest{Usagi: true},
		},
		"zero bool fields": {
			input: &boolTest{},
			want:  &boolTest{Usagi: false},
		},
		"bool ptr fields": {
			input: &boolPtrTest{Usagi: convertBoolPtr(true)},
			want:  &boolPtrTest{Usagi: convertBoolPtr(true)},
		},
		"nil bool ptr fields": {
			input: &boolPtrTest{},
			want:  &boolPtrTest{Usagi: (*bool)(nil)},
		},
		"map string to string fields": {
			input: &mapStringToStringTest{Usagi: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want:  &mapStringToStringTest{Usagi: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
		},
		"nil map string to string fields": {
			input: &mapStringToStringTest{},
			want:  &mapStringToStringTest{Usagi: (map[string]string)(nil)},
		},
		"map string to int fields": {
			input: &mapStringToIntTest{Usagi: map[string]int{"うさぎ": 20190122, "うさぎ2": 20190122, "うさぎ3": 20190122}},
			want:  &mapStringToIntTest{Usagi: map[string]int{"うさぎ": 20190122, "うさぎ2": 20190122, "うさぎ3": 20190122}},
		},
		"map string to float64 fields": {
			input: &mapStringToFloat64Test{Usagi: map[string]float64{"うさぎ": 20190122, "うさぎ2": 20190122, "うさぎ3": 20190122}},
			want:  &mapStringToFloat64Test{Usagi: map[string]float64{"うさぎ": 20190122, "うさぎ2": 20190122, "うさぎ3": 20190122}},
		},
		"map int to string fields": {
			input: &mapIntToStringTest{Usagi: map[int]string{201901221: "ハァ？", 201901222: "ウラ", 201901223: "フゥン"}},
			want:  &mapIntToStringTest{Usagi: map[int]string{201901221: "ハァ？", 201901222: "ウラ", 201901223: "フゥン"}},
		},
		"map int to int fields": {
			input: &mapIntToIntTest{Usagi: map[int]int{1: 201901221, 2: 201901222, 3: 201901223}},
			want:  &mapIntToIntTest{Usagi: map[int]int{1: 201901221, 2: 201901222, 3: 201901223}},
		},
		"map int to float64 fields": {
			input: &mapIntToFloat64Test{Usagi: map[int]float64{1: 201901221, 2: 201901222, 3: 201901223}},
			want:  &mapIntToFloat64Test{Usagi: map[int]float64{1: 201901221, 2: 201901222, 3: 201901223}},
		},
		"map struct to string fields": {
			input: &mapStructToStringTest{Usagi: map[stringTest]string{{Usagi: "ウサギ１"}: "ハァ？", {Usagi: "ウサギ２"}: "ウラ", {Usagi: "ウサギ３"}: "フゥン"}},
			want:  &mapStructToStringTest{Usagi: map[stringTest]string{{Usagi: "ウサギ１"}: "ハァ？", {Usagi: "ウサギ２"}: "ウラ", {Usagi: "ウサギ３"}: "フゥン"}},
		},
		"map struct to int fields": {
			input: &mapStructToIntTest{Usagi: map[stringTest]int{{Usagi: "ウサギ１"}: 201901221, {Usagi: "ウサギ２"}: 201901222, {Usagi: "ウサギ３"}: 201901223}},
			want:  &mapStructToIntTest{Usagi: map[stringTest]int{{Usagi: "ウサギ１"}: 201901221, {Usagi: "ウサギ２"}: 201901222, {Usagi: "ウサギ３"}: 201901223}},
		},
		"map struct to float64 fields": {
			input: &mapStructToFloat64Test{Usagi: map[stringTest]float64{{Usagi: "ウサギ１"}: 201901221, {Usagi: "ウサギ２"}: 201901222, {Usagi: "ウサギ３"}: 201901223}},
			want:  &mapStructToFloat64Test{Usagi: map[stringTest]float64{{Usagi: "ウサギ１"}: 201901221, {Usagi: "ウサギ２"}: 201901222, {Usagi: "ウサギ３"}: 201901223}},
		},
		"struct fields": {
			input: &structTest{
				StringTest:      stringTest{Usagi: "ヤハッ！"},
				StringSliceTest: stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
			},
			want: &structTest{
				StringTest:      stringTest{Usagi: "ヤハッ！"},
				StringSliceTest: stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
			},
		},
		"zero struct fields": {
			input: &structTest{},
			want: &structTest{
				StringTest:      stringTest{},
				StringSliceTest: stringSliceTest{},
			},
		},
		"struct slice fields": {
			input: &structSliceTest{
				SliceTest: []stringTest{
					{Usagi: "ハァ？"}, {Usagi: "ウラ"}, {Usagi: "フゥン"},
				},
			},
			want: &structSliceTest{
				SliceTest: []stringTest{{
					Usagi: "ハァ？"}, {Usagi: "ウラ"}, {Usagi: "フゥン"},
				},
			},
		},
		"nil struct slice fields": {
			input: &structSliceTest{},
			want:  &structSliceTest{SliceTest: ([]stringTest)(nil)},
		},
		"unexported fields": {
			input: &unexportedTest{usagi: "ヤハッ！"},
			want:  &unexportedTest{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer cleanup(t)
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskString(t *testing.T) {
	tests := map[string]struct {
		tag   string
		input string
		want  string
	}{
		"no tag": {
			tag:   "",
			input: "ヤハッ！",
			want:  "ヤハッ！",
		},
		"undefined tag": {
			tag:   "usagi!!",
			input: "ヤハッ！",
			want:  "ヤハッ！",
		},
		"filled": {
			tag:   MaskTypeFilled,
			input: "ヤハッ！",
			want:  "****",
		},
		"hide": {
			tag:   MaskTypeHide,
			input: "ヤハッ！",
			want:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer cleanup(t)
			got, err := MaskString(tt.tag, tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskInt(t *testing.T) {
	tests := map[string]struct {
		tag     string
		input   int
		want    int
		wantErr bool
	}{
		"no tag": {
			tag:   "",
			input: 20190122,
			want:  20190122,
		},
		"undefined tag": {
			tag:   "usagi!!",
			input: 20190122,
			want:  20190122,
		},
		"randomXX": {
			tag:     MaskTypeRandom + "XX",
			input:   20190122,
			wantErr: true,
		},
		"random30": {
			tag:   MaskTypeRandom + "30",
			input: 20190122,
			want:  9,
		},
		"random1000": {
			tag:   MaskTypeRandom + "1000",
			input: 20190122,
			want:  829,
		},
		"hide": {
			tag:   "hide",
			input: 0,
			want:  0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rand.Seed(rand.NewSource(1).Int63())
			defer cleanup(t)
			got, err := MaskInt(tt.tag, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestMaskFloat64(t *testing.T) {
	tests := map[string]struct {
		tag     string
		input   float64
		want    float64
		wantErr bool
	}{
		"no tag": {
			tag:   "",
			input: 20190122,
			want:  20190122,
		},
		"undefined tag": {
			tag:   "usagi!!",
			input: 20190122,
			want:  20190122,
		},
		"randomXX.4": {
			tag:     MaskTypeRandom + "XX.4",
			input:   20190122,
			wantErr: true,
		},
		"random4.XX": {
			tag:     MaskTypeRandom + "4.XX",
			input:   20190122,
			wantErr: true,
		},
		"random5.4": {
			tag:   MaskTypeRandom + "5.4",
			input: 20190122,
			want:  96011.8989,
		},
		"random1.1": {
			tag:   MaskTypeRandom + "1.1",
			input: 20190122,
			want:  9.6,
		},
		"random1": {
			tag:   MaskTypeRandom + "1",
			input: 20190122,
			want:  9.0,
		},
		"hide": {
			tag:   "hide",
			input: 20190122,
			want:  0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rand.Seed(rand.NewSource(1).Int63())
			defer cleanup(t)
			got, err := MaskFloat64(tt.tag, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestMaskFilled(t *testing.T) {
	type stringTest struct {
		Usagi string `mask:"filled"`
	}
	type stringMask5Test struct {
		Usagi string `mask:"filled5"`
	}
	type stringPtrTest struct {
		Usagi *string `mask:"filled"`
	}
	type stringPtrMask8Test struct {
		Usagi *string `mask:"filled8"`
	}
	type stringSliceTest struct {
		Usagi []string `mask:"filled"`
	}
	type stringSlicePtrTest struct {
		Usagi *[]string `mask:"filled"`
	}
	type stringToStringMapTest struct {
		Usagi map[string]string `mask:"filled"`
	}
	type intToStringMapTest struct {
		Usagi map[int]string `mask:"filled"`
	}
	type structToStringMapTest struct {
		Usagi map[stringTest]string `mask:"filled"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string": {
			input: "ヤハッ！",
			want:  "ヤハッ！",
		},
		"zero string": {
			input: "",
			want:  "",
		},
		"string ptr": {
			input: convertStringPtr("ヤハッ！"),
			want:  convertStringPtr("ヤハッ！"),
		},
		"nil string ptr": {
			input: (*string)(nil),
			want:  (*string)(nil),
		},
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！"},
			want:  &stringTest{Usagi: "****"},
		},
		"zero string fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
		},
		"string slice": {
			input: []string{"ハァ？", "ウラ", "フゥン"},
			want:  []string{"ハァ？", "ウラ", "フゥン"},
		},
		"nil string slice": {
			input: ([]string)(nil),
			want:  ([]string)(nil),
		},
		"string slice ptr": {
			input: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"}),
			want:  convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"}),
		},
		"nil string slice ptr": {
			input: (*[]string)(nil),
			want:  (*[]string)(nil),
		},
		"string map": {
			input: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"},
			want:  map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"},
		},
		"nil string map": {
			input: (map[string]string)(nil),
			want:  (map[string]string)(nil),
		},
		"string ptr fields": {
			input: &stringPtrTest{Usagi: convertStringPtr("ヤハッ！")},
			want:  &stringPtrTest{Usagi: convertStringPtr("****")},
		},
		"nil string ptr fields": {
			input: &stringPtrTest{},
			want:  &stringPtrTest{Usagi: (*string)(nil)},
		},
		"string slice fields": {
			input: &stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
			want:  &stringSliceTest{Usagi: []string{"***", "**", "***"}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"})},
			want:  &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"***", "**", "***"})},
		},
		"nil string slice ptr fields": {
			input: &stringSlicePtrTest{},
			want:  &stringSlicePtrTest{Usagi: (*[]string)(nil)},
		},
		"string to string map fields": {
			input: &stringToStringMapTest{Usagi: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want:  &stringToStringMapTest{Usagi: map[string]string{"うさぎ": "***", "うさぎ2": "**", "うさぎ3": "***"}},
		},
		"int to string map fields": {
			input: &intToStringMapTest{Usagi: map[int]string{1: "ハァ？", 2: "ウラ", 3: "フゥン"}},
			want:  &intToStringMapTest{Usagi: map[int]string{1: "***", 2: "**", 3: "***"}},
		},
		"struct to string map fields": {
			input: &structToStringMapTest{Usagi: map[stringTest]string{{Usagi: "ヤハッ！"}: "ハァ？", {Usagi: "ヤハッ！！"}: "ウラ", {Usagi: "ヤハッ！！！"}: "フゥン"}},
			want:  &structToStringMapTest{Usagi: map[stringTest]string{{Usagi: "ヤハッ！"}: "***", {Usagi: "ヤハッ！！"}: "**", {Usagi: "ヤハッ！！！"}: "***"}},
		},
		"filled 5 chars": {
			input: stringMask5Test{Usagi: "ヤハッ！"},
			want:  stringMask5Test{Usagi: "*****"},
		},
		"filled 8 chars": {
			input: stringPtrMask8Test{Usagi: convertStringPtr("ヤハッ！")},
			want:  stringPtrMask8Test{Usagi: convertStringPtr("********")},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer cleanup(t)
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskHashString(t *testing.T) {
	type stringTest struct {
		Usagi string `mask:"hash"`
	}
	type stringPtrTest struct {
		Usagi *string `mask:"hash"`
	}
	type stringSliceTest struct {
		Usagi []string `mask:"hash"`
	}
	type stringSlicePtrTest struct {
		Usagi *[]string `mask:"hash"`
	}
	type stringToStringMapTest struct {
		Usagi map[string]string `mask:"hash"`
	}
	type intToStringMapTest struct {
		Usagi map[int]string `mask:"hash"`
	}
	type structToStringMapTest struct {
		Usagi map[stringTest]string `mask:"hash"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string": {
			input: "ヤハッ！",
			want:  "ヤハッ！",
		},
		"zero string": {
			input: "",
			want:  "",
		},
		"string ptr": {
			input: convertStringPtr("ヤハッ！"),
			want:  convertStringPtr("ヤハッ！"),
		},
		"nil string ptr": {
			input: (*string)(nil),
			want:  (*string)(nil),
		},
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！"},
			want:  &stringTest{Usagi: "a6ab5728db57954641b2e155adc61f2cbdfc7063"},
		},
		"zero string fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
		},
		"string slice": {
			input: []string{"ハァ？", "ウラ", "フゥン"},
			want:  []string{"ハァ？", "ウラ", "フゥン"},
		},
		"nil string slice": {
			input: ([]string)(nil),
			want:  ([]string)(nil),
		},
		"string slice ptr": {
			input: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"}),
			want:  convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"}),
		},
		"nil string slice ptr": {
			input: (*[]string)(nil),
			want:  (*[]string)(nil),
		},
		"string map": {
			input: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"},
			want:  map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"},
		},
		"nil string map": {
			input: (map[string]string)(nil),
			want:  (map[string]string)(nil),
		},
		"string ptr fields": {
			input: &stringPtrTest{Usagi: convertStringPtr("ヤハッ！")},
			want:  &stringPtrTest{Usagi: convertStringPtr("a6ab5728db57954641b2e155adc61f2cbdfc7063")},
		},
		"nil string ptr fields": {
			input: &stringPtrTest{},
			want:  &stringPtrTest{Usagi: (*string)(nil)},
		},
		"string slice fields": {
			input: &stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
			want: &stringSliceTest{Usagi: []string{
				"48a8b33f36a35631f584844686adaba89a6f156a",
				"ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"})},
			want: &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{
				"48a8b33f36a35631f584844686adaba89a6f156a",
				"ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			})},
		},
		"nil string slice ptr fields": {
			input: &stringSlicePtrTest{},
			want:  &stringSlicePtrTest{Usagi: (*[]string)(nil)},
		},
		"string to string map fields": {
			input: &stringToStringMapTest{Usagi: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want: &stringToStringMapTest{Usagi: map[string]string{
				"うさぎ":  "48a8b33f36a35631f584844686adaba89a6f156a",
				"うさぎ2": "ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"うさぎ3": "17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"int to string map fields": {
			input: &intToStringMapTest{Usagi: map[int]string{1: "ハァ？", 2: "ウラ", 3: "フゥン"}},
			want: &intToStringMapTest{Usagi: map[int]string{
				1: "48a8b33f36a35631f584844686adaba89a6f156a",
				2: "ecef3e43f07f7150c089e99d5e1041259b1189d5",
				3: "17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"struct to string map fields": {
			input: &structToStringMapTest{Usagi: map[stringTest]string{{Usagi: "ヤハッ！"}: "ハァ？", {Usagi: "ヤハッ！！"}: "ウラ", {Usagi: "ヤハッ！！！"}: "フゥン"}},
			want: &structToStringMapTest{Usagi: map[stringTest]string{{
				Usagi: "ヤハッ！"}: "48a8b33f36a35631f584844686adaba89a6f156a",
				{Usagi: "ヤハッ！！"}:  "ecef3e43f07f7150c089e99d5e1041259b1189d5",
				{Usagi: "ヤハッ！！！"}: "17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer cleanup(t)
			got, err := Mask(tt.input)
			assert.Nil(t, err)

			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskRandom(t *testing.T) {
	type intTest struct {
		Usagi int `mask:"random1000"`
	}
	type int32Test struct {
		Usagi int32 `mask:"random1000"`
	}
	type int64Test struct {
		Usagi int64 `mask:"random1000"`
	}
	type intPtrTest struct {
		Usagi *int `mask:"random1000"`
	}
	type intSliceTest struct {
		Usagi []int `mask:"random1000"`
	}
	type int32SliceTest struct {
		Usagi []int32 `mask:"random1000"`
	}
	type int64SliceTest struct {
		Usagi []int64 `mask:"random1000"`
	}
	type intSlicePtrTest struct {
		Usagi *[]int `mask:"random1000"`
	}
	type float64Test struct {
		Usagi float64 `mask:"random5.4"`
	}
	type float64PtrTest struct {
		Usagi *float64 `mask:"random5.4"`
	}
	type float32SliceTest struct {
		Usagi []float32 `mask:"random5.4"`
	}
	type float64SliceTest struct {
		Usagi []float64 `mask:"random5.4"`
	}
	type float64SlicePtrTest struct {
		Usagi *[]float64 `mask:"random5.4"`
	}
	type stringToIntTest struct {
		Usagi map[string]int `mask:"random1000"`
	}
	type stringToInt32Test struct {
		Usagi map[string]int32 `mask:"random1000"`
	}
	type stringToInt64Test struct {
		Usagi map[string]int64 `mask:"random1000"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"int fields": {
			input: &intTest{Usagi: 20190122},
			want:  &intTest{Usagi: 829},
		},
		"int32 fields": {
			input: &int32Test{Usagi: 20190122},
			want:  &int32Test{Usagi: 829},
		},
		"int64 fields": {
			input: &int64Test{Usagi: 20190122},
			want:  &int64Test{Usagi: 829},
		},
		"zero int fields": {
			input: &intTest{},
			want:  &intTest{Usagi: 0},
		},
		"int ptr fields": {
			input: &intPtrTest{Usagi: convertIntPtr(20190122)},
			want:  &intPtrTest{Usagi: convertIntPtr(829)},
		},
		"nil int ptr fields": {
			input: &intPtrTest{},
			want:  &intPtrTest{Usagi: nil},
		},
		"int slice fields": {
			input: &intSliceTest{Usagi: []int{20190122, 20200501, 20200501}},
			want:  &intSliceTest{Usagi: []int{829, 830, 400}},
		},
		"int32 slice fields": {
			input: &int32SliceTest{Usagi: []int32{20190122, 20200501, 20200501}},
			want:  &int32SliceTest{Usagi: []int32{829, 830, 400}},
		},
		"int64 slice fields": {
			input: &int64SliceTest{Usagi: []int64{20190122, 20200501, 20200501}},
			want:  &int64SliceTest{Usagi: []int64{829, 830, 400}},
		},
		"nil int slice fields": {
			input: &intSliceTest{},
			want:  &intSliceTest{Usagi: ([]int)(nil)},
		},
		"int slice ptr fields": {
			input: &intSlicePtrTest{Usagi: convertIntSlicePtr([]int{20190122, 20200501, 20200501})},
			want:  &intSlicePtrTest{Usagi: convertIntSlicePtr([]int{829, 830, 400})},
		},
		"nil int slice ptr fields": {
			input: &intSlicePtrTest{},
			want:  &intSlicePtrTest{Usagi: (*[]int)(nil)},
		},
		"float64 fields": {
			input: &float64Test{Usagi: 20190122},
			want:  &float64Test{Usagi: 96011.8989},
		},
		"zero float64 fields": {
			input: &float64Test{},
			want:  &float64Test{Usagi: 0},
		},
		"float64 ptr fields": {
			input: &float64PtrTest{Usagi: convertFloat64Ptr(20190122)},
			want:  &float64PtrTest{Usagi: convertFloat64Ptr(96011.8989)},
		},
		"nil float64 ptr fields": {
			input: &float64PtrTest{},
			want:  &float64PtrTest{Usagi: nil},
		},
		"float32 slice fields": {
			input: &float32SliceTest{Usagi: []float32{20190122, 20200501, 20200501}},
			want:  &float32SliceTest{Usagi: []float32{96011.8989, 90863.3149, 32310.0201}},
		},
		"float64 slice fields": {
			input: &float64SliceTest{Usagi: []float64{20190122, 20200501, 20200501}},
			want:  &float64SliceTest{Usagi: []float64{96011.8989, 90863.3149, 32310.0201}},
		},
		"nil float64 slice fields": {
			input: &float64SliceTest{},
			want:  &float64SliceTest{Usagi: ([]float64)(nil)},
		},
		"float64 slice ptr fields": {
			input: &float64SlicePtrTest{Usagi: convertFloat64SlicePtr([]float64{20190122, 20200501, 20200501})},
			want:  &float64SlicePtrTest{Usagi: convertFloat64SlicePtr([]float64{96011.8989, 90863.3149, 32310.0201})},
		},
		"nil float64 slice ptr fields": {
			input: &float64SlicePtrTest{},
			want:  &float64SlicePtrTest{Usagi: (*[]float64)(nil)},
		},
		"string to int map fields": {
			input: &stringToIntTest{Usagi: map[string]int{"うさぎ": 20190122, "ちいかわ": 20200501, "はちわれ": 20200501}},
			want:  &stringToIntTest{Usagi: map[string]int{"はちわれ": 400, "うさぎ": 829, "ちいかわ": 830}},
		},
		"string to int32 map fields": {
			input: &stringToInt32Test{Usagi: map[string]int32{"うさぎ": 20190122, "ちいかわ": 20200501, "はちわれ": 20200501}},
			want:  &stringToInt32Test{Usagi: map[string]int32{"はちわれ": 400, "うさぎ": 829, "ちいかわ": 830}},
		},
		"string to int64 map fields": {
			input: &stringToInt64Test{Usagi: map[string]int64{"うさぎ": 20190122, "ちいかわ": 20200501, "はちわれ": 20200501}},
			want:  &stringToInt64Test{Usagi: map[string]int64{"はちわれ": 400, "うさぎ": 829, "ちいかわ": 830}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer cleanup(t)
			rand.Seed(rand.NewSource(1).Int63())
			got, err := Mask(tt.input)
			if assert.NoError(t, err) {
				if diff := cmp.Diff(tt.want, got, cmpopts.SortMaps(func(i, j string) bool { return i < j })); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestMaskHide(t *testing.T) {
	type stringTest struct {
		Usagi string `mask:"hide"`
	}
	type stringPtrTest struct {
		Usagi *string `mask:"hide"`
	}
	type stringSliceTest struct {
		Usagi []string `mask:"hide"`
	}
	type stringSlicePtrTest struct {
		Usagi *[]string `mask:"hide"`
	}
	type intTest struct {
		Usagi int `mask:"hide"`
	}
	type float64Test struct {
		Usagi float64 `mask:"hide"`
	}
	type boolTest struct {
		Usagi bool `mask:"hide"`
	}
	type mapStringToStringTest struct {
		Usagi map[string]string `mask:"hide"`
	}
	type structTest struct {
		StringTest stringTest `mask:"hide"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！"},
			want:  &stringTest{Usagi: ""},
		},
		"string empty fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
		},
		"string ptr fields": {
			input: &stringPtrTest{Usagi: convertStringPtr("ヤハッ！")},
			want:  &stringPtrTest{},
		},
		"nil string ptr fields": {
			input: &stringPtrTest{},
			want:  &stringPtrTest{Usagi: nil},
		},
		"string slice fields": {
			input: &stringSliceTest{Usagi: []string{"ハァ？", "ウラ", "フゥン"}},
			want:  &stringSliceTest{},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"})},
			want:  &stringSlicePtrTest{},
		},
		"nil string slice ptr fields": {
			input: &stringSlicePtrTest{},
			want:  &stringSlicePtrTest{Usagi: (*[]string)(nil)},
		},
		"int fields": {
			input: &intTest{Usagi: 20190122},
			want:  &intTest{Usagi: 0},
		},
		"zero int fields": {
			input: &intTest{},
			want:  &intTest{Usagi: 0},
		},
		"float64 fields": {
			input: &float64Test{Usagi: 20190122},
			want:  &float64Test{Usagi: 0},
		},
		"zero float64 fields": {
			input: &float64Test{},
			want:  &float64Test{Usagi: 0},
		},
		"bool fields": {
			input: &boolTest{Usagi: true},
			want:  &boolTest{Usagi: false},
		},
		"zero bool fields": {
			input: &boolTest{},
			want:  &boolTest{},
		},
		"map string to string fields": {
			input: &mapStringToStringTest{Usagi: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want:  &mapStringToStringTest{},
		},
		"nil map string to string fields": {
			input: &mapStringToStringTest{},
			want:  &mapStringToStringTest{},
		},
		"struct fields": {
			input: &structTest{
				StringTest: stringTest{Usagi: "ヤハッ！"},
			},
			want: &structTest{
				StringTest: stringTest{},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer cleanup(t)
			rand.Seed(rand.NewSource(1).Int63())
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func allowUnexported(v any) cmp.Options {
	var options cmp.Options
	if !reflect.ValueOf(v).IsValid() {
		return options
	}
	rt, ok := getStructType(reflect.TypeOf(v))
	if !ok {
		return options
	}

	rv := reflect.New(rt).Elem()
	options = append(options, cmp.AllowUnexported(rv.Interface()))
	for i := 0; i < rv.NumField(); i++ {
		if rt2, ok := getStructType(rv.Field(i).Type()); ok {
			rv2 := reflect.New(rt2).Elem()
			options = append(options, allowUnexported(rv2.Interface())...)
		}
	}

	return options
}

func getStructType(rt reflect.Type) (reflect.Type, bool) {
	switch rt.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Slice:
		return getStructType(rt.Elem())
	case reflect.Struct:
		return rt, true
	default:
		return rt, false
	}
}

func convertStringPtr(s string) *string {
	return &s
}
func convertStringSlicePtr(s []string) *[]string {
	return &s
}
func convertIntPtr(i int) *int {
	return &i
}
func convertIntSlicePtr(i []int) *[]int {
	return &i
}
func convertFloat64Ptr(f float64) *float64 {
	return &f
}
func convertFloat64SlicePtr(f []float64) *[]float64 {
	return &f
}
func convertBoolPtr(v bool) *bool {
	return &v
}

func cleanup(t *testing.T) {
	t.Helper()
	typeToStructMap.Range(func(key, _ any) bool {
		typeToStructMap.Delete(key)
		return false
	})
}

// MaskRegExp is sample to add custom mask function
func MaskRegExp(arg, value string) (string, error) {
	var (
		reg *regexp.Regexp
		err error
	)
	reg, err = regexp.Compile(arg)
	if err != nil {
		return "", err
	}

	indexes := reg.FindStringSubmatchIndex(value)
	if len(indexes) >= 4 && indexes[2] >= 0 && indexes[3] >= 0 {
		var sb strings.Builder
		sb.WriteString(value[:indexes[2]])
		sb.WriteString(strings.Repeat(maskChar, utf8.RuneCountInString(value[indexes[2]:indexes[3]])))
		sb.WriteString(value[indexes[3]:])
		return sb.String(), nil
	}

	return value, nil
}

func ExampleMaskRegExp() {
	maskTypeRegExp := "regexp"
	RegisterMaskStringFunc(maskTypeRegExp, MaskRegExp)

	type RegExpTest struct {
		Usagi string `mask:"regexp(ヤハ)*"`
	}

	input := RegExpTest{Usagi: "ヤハッ！"}
	got, _ := Mask(input)
	fmt.Printf("Usagi %s\n", got.(RegExpTest).Usagi)

	// Output:
	// Usagi **ッ！
}

type benchStruct2 struct {
	Case1 string
	Case2 int
	Case3 bool
	Case4 []string
	Case5 map[string]string
	Case6 map[int]string
}
type benchStruct1 struct {
	Case1  string
	Case11 string
	Case2  int
	Case3  bool
	Case4  []string
	Case44 []string
	Case5  map[string]string
	Case6  *benchStruct2   `mask:"struct"`
	Case7  []*benchStruct2 `mask:"struct"`
}

func createChiikawa(s string) *benchStruct2 {
	return &benchStruct2{
		Case1: s,
		Case2: 20200501,
		Case3: false,
		Case4: []string{
			"わァ………",
			"ァ…………",
			"ァ…………ゥ…………",
		},
		Case5: map[string]string{
			"ヤーッ！":     "ヤーッ！！！",
			"ヤーッ！！":    "ヤーッ！！！",
			"ヤーッ！！！":   "ヤーッ！！！",
			"ヤーッ！！！！":  "ヤーッ！！！",
			"ヤーッ！！！！！": "ヤーッ！！！",
		},
		Case6: map[int]string{
			1: "ヤーッ！！！",
			2: "ヤーッ！！！",
			3: "ヤーッ！！！",
			4: "ヤーッ！！！",
			5: "ヤーッ！！！",
		},
	}
}

func createHachiware() *benchStruct1 {
	return &benchStruct1{
		Case1:  "はちわれ",
		Case11: "はちわれ",
		Case2:  20200501,
		Case3:  true,
		Case4: []string{
			"もしかして",
			"ベンチマーク",
			"とってる…",
			"ってコト！？",
		},
		Case44: []string{
			"もしかして",
			"ベンチマーク",
			"とってる…",
			"ってコト！？",
		},
		Case5: map[string]string{
			"モモンガ": "慰めろッ",
			"はちわれ": "なになに！？",
		},
		Case6: createChiikawa("ちいかわ"),
		Case7: []*benchStruct2{
			createChiikawa("ちいかわ1"),
			createChiikawa("ちいかわ2"),
			createChiikawa("ちいかわ3"),
			createChiikawa("ちいかわ4"),
		},
	}
}

func BenchmarkMask(b *testing.B) {
	hachiware := createChiikawa("hoge")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Mask(hachiware)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGoMasker(b *testing.B) {
	hachiware := createChiikawa("hoge")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := masker.Struct(hachiware)
		if err != nil {
			b.Error(err)
		}
	}
}
