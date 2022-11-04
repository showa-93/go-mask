package maskgo

import (
	"math/rand"
	"regexp"
	"testing"

	"github.com/goccy/go-reflect"

	"github.com/ggwhite/go-masker"
	"github.com/google/go-cmp/cmp"
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
	type structTest struct {
		StringTest      stringTest
		StringSliceTest stringSliceTest
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
			tag:   "filled!!",
			input: "ヤハッ！",
			want:  "****",
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

func TestMaskFilled(t *testing.T) {
	type stringTest struct {
		Usagi string `mask:"filled"`
	}
	type stringPtrTest struct {
		Usagi *string `mask:"filled"`
	}
	type stringSliceTest struct {
		Usagi []string `mask:"filled"`
	}
	type stringSlicePtrTest struct {
		Usagi *[]string `mask:"filled"`
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
			want:  &stringSliceTest{Usagi: []string{"48a8b33f36a35631f584844686adaba89a6f156a", "ecef3e43f07f7150c089e99d5e1041259b1189d5", "17fa078ad3f2c34c17ee58b9119963548ddcf1ef"}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"ハァ？", "ウラ", "フゥン"})},
			want:  &stringSlicePtrTest{Usagi: convertStringSlicePtr([]string{"48a8b33f36a35631f584844686adaba89a6f156a", "ecef3e43f07f7150c089e99d5e1041259b1189d5", "17fa078ad3f2c34c17ee58b9119963548ddcf1ef"})},
		},
		"nil string slice ptr fields": {
			input: &stringSlicePtrTest{},
			want:  &stringSlicePtrTest{Usagi: (*[]string)(nil)},
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

func TestMaskInt(t *testing.T) {
	tests := map[string]struct {
		tag   string
		input int
		want  int
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
		"random30": {
			tag:   "random30",
			input: 20190122,
			want:  9,
		},
		"random1000": {
			tag:   "random1000",
			input: 20190122,
			want:  829,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rand.Seed(rand.NewSource(1).Int63())
			defer cleanup(t)
			got, err := MaskInt(tt.tag, tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskFloat64(t *testing.T) {
	tests := map[string]struct {
		tag   string
		input float64
		want  float64
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
		"random5.4": {
			tag:   "random5.4",
			input: 20190122,
			want:  96011.8989,
		},
		"random1.0": {
			tag:   "random1.0",
			input: 20190122,
			want:  9.0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rand.Seed(rand.NewSource(1).Int63())
			defer cleanup(t)
			got, err := MaskFloat64(tt.tag, tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskRandom(t *testing.T) {
	type intTest struct {
		Usagi int `mask:"random1000"`
	}
	type intPtrTest struct {
		Usagi *int `mask:"random1000"`
	}
	type intSliceTest struct {
		Usagi []int `mask:"random1000"`
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
	type float64SliceTest struct {
		Usagi []float64 `mask:"random5.4"`
	}
	type float64SlicePtrTest struct {
		Usagi *[]float64 `mask:"random5.4"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"int fields": {
			input: &intTest{Usagi: 20190122},
			want:  &intTest{Usagi: 829},
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

func validSha1(s string) bool {
	ok, _ := regexp.MatchString("^[a-fA-F0-9]{40}$", s)
	return ok
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
	typeToStruct.Range(func(key, _ any) bool {
		typeToStruct.Delete(key)
		return false
	})
}

type benchStruct2 struct {
	Case1 string
	Case2 int `mask:"random1000"`
	Case3 bool
	Case4 []string
	Case5 map[string]string
}
type benchStruct1 struct {
	Case1 string `mask:"name"`
	Case2 int
	Case3 bool
	Case4 []string `mask:"name"`
	Case5 map[string]string
	Case6 *benchStruct2   `mask:"struct"`
	Case7 []*benchStruct2 `mask:"struct"`
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
	}
}

func BenchmarkMask(b *testing.B) {
	hachiware := benchStruct1{
		Case1: "はちわれ",
		Case2: 20200501,
		Case3: true,
		Case4: []string{
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Mask(&hachiware); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGoMasker(b *testing.B) {
	hachiware := benchStruct1{
		Case1: "はちわれ",
		Case2: 20200501,
		Case3: true,
		Case4: []string{
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := masker.Struct(&hachiware); err != nil {
			b.Error(err)
		}
	}
}
