package mask

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func Example() {
	rand.Seed(12345)
	type Address struct {
		PostCode string `mask:"zero"`
	}
	type User struct {
		ID      string
		Name    string `mask:"filled"`
		Age     int    `mask:"random100"`
		Address Address
	}

	user := User{
		ID:   "123456",
		Name: "Usagi",
		Age:  3,
		Address: Address{
			PostCode: "123-4567",
		},
	}
	maskUser, err := Mask(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v", maskUser)

	// Output:
	// {ID:123456 Name:***** Age:83 Address:{PostCode:}}
}

func ExampleRegisterMaskField() {
	rand.Seed(12345)
	type User2 struct {
		ID      string
		Name    string
		Age     int
		ExtData map[string]string
	}
	user := User2{
		ID:   "123456",
		Name: "Usagi",
		Age:  3,
		ExtData: map[string]string{
			"ID":       "123456",
			"Favorite": "Cat",
		},
	}

	RegisterMaskField("ID", "zero")
	RegisterMaskField("Age", "random100")
	RegisterMaskField("Name", "filled4")
	RegisterMaskField("Favorite", "filled6")
	maskUser, err := Mask(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v", maskUser)

	// Output:
	// {ID: Name:**** Age:83 ExtData:map[Favorite:****** ID:]}
}

func BenchmarkMask(b *testing.B) {
	type BenchTarget2 struct {
		I  int       `mask:"random100"`
		S  string    `mask:"fixed"`
		SS []string  `mask:"filled"`
		IS []int     `mask:"rondom100"`
		FS []float64 `mask:"rondom100"`
	}

	type BenchTarget struct {
		I  int    `mask:"zero"`
		S  string `mask:"filled"`
		M  map[string]string
		SS []string  `mask:"filled"`
		IS []int     `mask:"rondom100"`
		FS []float64 `mask:"rondom100"`
		B  *BenchTarget2
	}

	RegisterMaskField("Hoge", MaskTypeFixed)
	RegisterMaskField("Bob", MaskTypeFilled+"4")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v := BenchTarget{
			I: 1,
			S: "Hello World",
			M: map[string]string{
				"Hoge": "Fuga",
				"Bob":  "Alica",
			},
			SS: []string{
				"One",
				"Two",
				"Three",
			},
			IS: []int{
				1,
				2,
				3,
			},
			FS: []float64{
				1.0,
				2.0,
				3.0,
			},
			B: &BenchTarget2{
				I: 2,
				S: "Hello World2",
				SS: []string{
					"One",
					"Two",
					"Three",
				},
				IS: []int{
					1,
					2,
					3,
				},
				FS: []float64{
					1,
					2,
					3,
				},
			},
		}
		Mask(v)
	}
}

func TestMask(t *testing.T) {
	tests := map[string]struct {
		prepare func(*Masker)
		input   any
		want    any
		isErr   bool
	}{
		"string": {
			prepare: func(*Masker) {},
			input:   "サンクチュアリ",
			want:    "サンクチュアリ",
		},
		"int": {
			prepare: func(*Masker) {},
			input:   int(100),
			want:    int(100),
		},
		"uint": {
			prepare: func(*Masker) {},
			input:   uint(100),
			want:    uint(100),
		},
		"float64": {
			prepare: func(*Masker) {},
			input:   float64(100.12),
			want:    float64(100.12),
		},
		"complex128": {
			prepare: func(*Masker) {},
			input:   complex128(100 + 12i),
			want:    complex128(100 + 12i),
		},
		"byte": {
			prepare: func(*Masker) {},
			input:   byte(2),
			want:    byte(2),
		},
		"struct string": {
			prepare: func(m *Masker) {
				RegisterMaskStringFunc("test", func(arg, value string) (string, error) {
					return "test", nil
				})
			},
			input: struct {
				String string `mask:"test"`
			}{"チャス"},
			want: struct {
				String string `mask:"test"`
			}{"test"},
		},
		"struct int": {
			prepare: func(m *Masker) {
				RegisterMaskIntFunc("test", func(arg string, value int) (int, error) {
					return math.MaxInt, nil
				})
			},
			input: struct {
				Int int `mask:"test"`
			}{1234},
			want: struct {
				Int int `mask:"test"`
			}{math.MaxInt},
		},
		"struct uint": {
			prepare: func(m *Masker) {
				RegisterMaskUintFunc("test", func(arg string, value uint) (uint, error) {
					return math.MaxUint, nil
				})
			},
			input: struct {
				Uint uint `mask:"test"`
			}{1234},
			want: struct {
				Uint uint `mask:"test"`
			}{math.MaxUint},
		},
		"struct float64": {
			prepare: func(m *Masker) {
				RegisterMaskFloat64Func("test", func(arg string, value float64) (float64, error) {
					return math.MaxFloat64, nil
				})
			},
			input: struct {
				Float64 float64 `mask:"test"`
			}{1234.5678},
			want: struct {
				Float64 float64 `mask:"test"`
			}{math.MaxFloat64},
		},
		"struct with private field": {
			prepare: func(m *Masker) {
				RegisterMaskFloat64Func("test", func(arg string, value float64) (float64, error) {
					return math.MaxFloat64, nil
				})
			},
			input: struct {
				Float64 float64 `mask:"test"`
				private string  `mask:"test"`
			}{
				Float64: 1234.5678,
				private: "x",
			},
			want: struct {
				Float64 float64 `mask:"test"`
				private string  `mask:"test"`
			}{Float64: math.MaxFloat64},
		},
	}

	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				defer cleanup(t)
				defaultMasker.Cache(cache)
				tt.prepare(defaultMasker)
				got, err := Mask(tt.input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_Primitive(t *testing.T) {
	type Tag struct {
		String     string     `mask:"test"`
		Int        int        `mask:"test"`
		Int8       int8       `mask:"test"`
		Int16      int16      `mask:"test"`
		Int32      int32      `mask:"test"`
		Int64      int64      `mask:"test"`
		Uint       uint       `mask:"test"`
		Uint8      uint8      `mask:"test"`
		Uint16     uint16     `mask:"test"`
		Uint32     uint32     `mask:"test"`
		Uint64     uint64     `mask:"test"`
		Float32    float32    `mask:"test"`
		Float64    float64    `mask:"test"`
		Complex64  complex64  `mask:"test"`
		Complex128 complex128 `mask:"test"`
		Byte       byte       `mask:"test"`
	}
	type NoTag struct {
		String     string
		Int        int
		Int8       int8
		Int16      int16
		Int32      int32
		Int64      int64
		Uint       uint
		Uint8      uint8
		Uint16     uint16
		Uint32     uint32
		Uint64     uint64
		Float32    float32
		Float64    float64
		Complex64  complex64
		Complex128 complex128
		Byte       byte
	}
	type Test struct {
		Tag
		NoTag
	}
	input := Test{
		Tag: Tag{
			String:     "サンクチュアリ -聖域-",
			Int:        1000,
			Int8:       12,
			Int16:      2000,
			Int32:      3000,
			Int64:      4000,
			Uint:       5000,
			Uint8:      12,
			Uint16:     6000,
			Uint32:     7000,
			Uint64:     8000,
			Float32:    123.456,
			Float64:    654.321,
			Complex64:  (1234 + 10i),
			Complex128: (4321 + 20i),
			Byte:       2,
		},
		NoTag: NoTag{
			String:     "サンクチュアリ -聖域-",
			Int:        1000,
			Int8:       12,
			Int16:      2000,
			Int32:      3000,
			Int64:      4000,
			Uint:       5000,
			Uint8:      12,
			Uint16:     6000,
			Uint32:     7000,
			Uint64:     8000,
			Float32:    123.456,
			Float64:    654.321,
			Complex64:  (1234 + 10i),
			Complex128: (4321 + 20i),
			Byte:       2,
		},
	}
	tests := map[string]struct {
		prepare func(*Masker)
		want    Test
		isErr   bool
	}{
		"no masking functions": {
			prepare: func(m *Masker) {},
			want:    input,
		},
		"register masking functions": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "test")
			},
			want: Test{
				Tag: Tag{
					String:     "test",
					Int:        math.MaxInt,
					Int8:       -1, // overflow
					Int16:      -1, // overflow
					Int32:      -1, // overflow
					Int64:      math.MaxInt64,
					Uint:       math.MaxUint,
					Uint8:      math.MaxUint8,  // overflow
					Uint16:     math.MaxUint16, // overflow
					Uint32:     math.MaxUint32, // overflow
					Uint64:     math.MaxUint64,
					Float32:    float32(math.Inf(0)), // overflow
					Float64:    math.MaxFloat64,
					Complex64:  (1234 + 10i),
					Complex128: (4321 + 20i),
					Byte:       255, // overflow
				},
				NoTag: input.NoTag,
			},
		},
		"register mask field name": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "field")
				m.RegisterMaskField("String", "field")
				m.RegisterMaskField("Int", "field")
				m.RegisterMaskField("Int8", "field")
				m.RegisterMaskField("Int16", "field")
				m.RegisterMaskField("Int32", "field")
				m.RegisterMaskField("Int64", "field")
				m.RegisterMaskField("Uint", "field")
				m.RegisterMaskField("Uint8", "field")
				m.RegisterMaskField("Uint16", "field")
				m.RegisterMaskField("Uint32", "field")
				m.RegisterMaskField("Uint64", "field")
				m.RegisterMaskField("Float32", "field")
				m.RegisterMaskField("Float64", "field")
				m.RegisterMaskField("Complex64", "field")
				m.RegisterMaskField("Complex128", "field")
				m.RegisterMaskField("Byte", "field")
			},
			want: Test{
				Tag: input.Tag,
				NoTag: NoTag{
					String:     "test",
					Int:        math.MaxInt,
					Int8:       -1, // overflow
					Int16:      -1, // overflow
					Int32:      -1, // overflow
					Int64:      math.MaxInt64,
					Uint:       math.MaxUint,
					Uint8:      math.MaxUint8,  // overflow
					Uint16:     math.MaxUint16, // overflow
					Uint32:     math.MaxUint32, // overflow
					Uint64:     math.MaxUint64,
					Float32:    float32(math.Inf(0)), // overflow
					Float64:    math.MaxFloat64,
					Complex64:  (1234 + 10i),
					Complex128: (4321 + 20i),
					Byte:       255, // overflow
				},
			},
		},
	}

	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				m := NewMasker()
				m.Cache(cache)
				tt.prepare(m)
				got, err := m.Mask(input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_Array(t *testing.T) {
	type Struct1 struct {
		String string `mask:"test"`
	}
	type Struct2 struct {
		String string
	}
	type CustomArrayOfStruct [3]Struct1
	type CustomArrayOfStruct2 [3]Struct2
	type CustomArrayOfByte [3]byte

	type Tag struct {
		String         [3]string           `mask:"test"`
		Int            [3]int              `mask:"test"`
		Uint           [3]uint             `mask:"test"`
		Float64        [3]float64          `mask:"test"`
		Complex128     [3]complex128       `mask:"test"`
		Byte           [3]byte             `mask:"test"`
		CustomOfByte   CustomArrayOfByte   `mask:"test"`
		Struct         [3]Struct1          `mask:"test"`
		CustomOfStruct CustomArrayOfStruct `mask:"test"`
	}
	type NoTag struct {
		String         [3]string
		Int            [3]int
		Uint           [3]uint
		Float64        [3]float64
		Complex128     [3]complex128
		Byte           [3]byte
		CustomOfByte   CustomArrayOfByte
		Struct         [3]Struct2
		CustomOfStruct CustomArrayOfStruct2
	}
	type Test struct {
		Tag
		NoTag
	}
	input := Test{
		Tag: Tag{
			String:         [3]string{"猿将", "猿谷", "猿桜"},
			Int:            [3]int{-1, 10, 100},
			Uint:           [3]uint{1, 2, 3},
			Float64:        [3]float64{1.1, 1000.123, 999.0},
			Complex128:     [3]complex128{100 + 1i, 10i, 10},
			Byte:           [3]byte{1, 2, 3},
			CustomOfByte:   CustomArrayOfByte{1, 2, 3},
			Struct:         [3]Struct1{{"猿空"}, {"猿岳"}, {"猿河"}},
			CustomOfStruct: CustomArrayOfStruct{{"猿空"}, {"猿岳"}, {"猿河"}},
		},
		NoTag: NoTag{
			String:         [3]string{"猿将", "猿谷", "猿桜"},
			Int:            [3]int{-1, 10, 100},
			Uint:           [3]uint{1, 2, 3},
			Float64:        [3]float64{1.1, 1000.123, 999.0},
			Complex128:     [3]complex128{100 + 1i, 10i, 10},
			Byte:           [3]byte{1, 2, 3},
			CustomOfByte:   CustomArrayOfByte{1, 2, 3},
			Struct:         [3]Struct2{{"猿空"}, {"猿岳"}, {"猿河"}},
			CustomOfStruct: CustomArrayOfStruct2{{"猿空"}, {"猿岳"}, {"猿河"}},
		},
	}
	tests := map[string]struct {
		prepare func(*Masker)
		want    Test
		isErr   bool
	}{
		"no masking functions": {
			prepare: func(m *Masker) {},
			want:    input,
		},
		"register masking functions": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "test")
			},
			want: Test{
				Tag: Tag{
					String:         [3]string{"test", "test", "test"},
					Int:            [3]int{math.MaxInt, math.MaxInt, math.MaxInt},
					Uint:           [3]uint{math.MaxUint, math.MaxUint, math.MaxUint},
					Float64:        [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
					Complex128:     [3]complex128{100 + 1i, 10i, 10},
					Byte:           [3]byte{255, 255, 255},
					CustomOfByte:   CustomArrayOfByte{255, 255, 255},
					Struct:         [3]Struct1{{"test"}, {"test"}, {"test"}},
					CustomOfStruct: CustomArrayOfStruct{{"test"}, {"test"}, {"test"}},
				},
				NoTag: input.NoTag,
			},
		},
		"register mask field name": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "field")
				m.RegisterMaskField("String", "field")
				m.RegisterMaskField("Int", "field")
				m.RegisterMaskField("Int8", "field")
				m.RegisterMaskField("Int16", "field")
				m.RegisterMaskField("Int32", "field")
				m.RegisterMaskField("Int64", "field")
				m.RegisterMaskField("Uint", "field")
				m.RegisterMaskField("Uint8", "field")
				m.RegisterMaskField("Uint16", "field")
				m.RegisterMaskField("Uint32", "field")
				m.RegisterMaskField("Uint64", "field")
				m.RegisterMaskField("Float32", "field")
				m.RegisterMaskField("Float64", "field")
				m.RegisterMaskField("Complex64", "field")
				m.RegisterMaskField("Complex128", "field")
				m.RegisterMaskField("Byte", "field")
				m.RegisterMaskField("Struct", "field")
				m.RegisterMaskField("CustomOfByte", "field")
				m.RegisterMaskField("CustomOfStruct", "field")
			},
			want: Test{
				Tag: input.Tag,
				NoTag: NoTag{
					String:         [3]string{"test", "test", "test"},
					Int:            [3]int{math.MaxInt, math.MaxInt, math.MaxInt},
					Uint:           [3]uint{math.MaxUint, math.MaxUint, math.MaxUint},
					Float64:        [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
					Complex128:     [3]complex128{100 + 1i, 10i, 10},
					Byte:           [3]byte{255, 255, 255},
					CustomOfByte:   CustomArrayOfByte{255, 255, 255},
					Struct:         [3]Struct2{{"test"}, {"test"}, {"test"}},
					CustomOfStruct: CustomArrayOfStruct2{{"test"}, {"test"}, {"test"}},
				},
			},
		},
	}

	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				m := NewMasker()
				tt.prepare(m)
				got, err := m.Mask(input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_Slice(t *testing.T) {
	type Struct1 struct {
		String string `mask:"test"`
	}
	type Struct2 struct {
		String string
	}
	type customString []string
	type customInt []int
	type customUint []uint
	type customFloat64 []float64
	type customComplex128 []complex128
	type customByte []byte
	type customStruct []Struct1
	type customStruct2 []Struct2

	type Tag struct {
		String     []string     `mask:"test"`
		Int        []int        `mask:"test"`
		Uint       []uint       `mask:"test"`
		Float64    []float64    `mask:"test"`
		Complex128 []complex128 `mask:"test"`
		Byte       []byte       `mask:"test"`
		Struct     []Struct1    `mask:"test"`

		CustomString     customString     `mask:"test"`
		CustomInt        customInt        `mask:"test"`
		CustomUint       customUint       `mask:"test"`
		CustomFloat64    customFloat64    `mask:"test"`
		CustomComplex128 customComplex128 `mask:"test"`
		CustomByte       customByte       `mask:"test"`
		CustomStruct     customStruct     `mask:"test"`

		ZeroGuard string
	}
	type NoTag struct {
		String     []string
		Int        []int
		Uint       []uint
		Float64    []float64
		Complex128 []complex128
		Byte       []byte
		Struct     []Struct2
		ZeroGuard  string

		CustomString     customString
		CustomInt        customInt
		CustomUint       customUint
		CustomFloat64    customFloat64
		CustomComplex128 customComplex128
		CustomByte       customByte
		CustomStruct     customStruct2
	}
	type Test struct {
		Tag
		NoTag
	}
	input := Test{
		Tag: Tag{
			String:     []string{"猿将", "猿谷", "猿桜"},
			Int:        []int{-1, 10, 100},
			Uint:       []uint{1, 2, 3},
			Float64:    []float64{1.1, 1000.123, 999.0},
			Complex128: []complex128{100 + 1i, 10i, 10},
			Byte:       []byte{1, 2, 3},
			Struct:     []Struct1{{"猿空"}, {"猿岳"}, {"猿河"}},

			CustomString:     customString{"猿将", "猿谷", "猿桜"},
			CustomInt:        customInt{-1, 10, 100},
			CustomUint:       customUint{1, 2, 3},
			CustomFloat64:    customFloat64{1.1, 1000.123, 999.0},
			CustomComplex128: customComplex128{100 + 1i, 10i, 10},
			CustomByte:       customByte{1, 2, 3},
			CustomStruct:     customStruct{{"猿空"}, {"猿岳"}, {"猿河"}},
		},
		NoTag: NoTag{
			String:     []string{"猿将", "猿谷", "猿桜"},
			Int:        []int{-1, 10, 100},
			Uint:       []uint{1, 2, 3},
			Float64:    []float64{1.1, 1000.123, 999.0},
			Complex128: []complex128{100 + 1i, 10i, 10},
			Byte:       []byte{1, 2, 3},
			Struct:     []Struct2{{"猿空"}, {"猿岳"}, {"猿河"}},

			CustomString:     customString{"猿将", "猿谷", "猿桜"},
			CustomInt:        customInt{-1, 10, 100},
			CustomUint:       customUint{1, 2, 3},
			CustomFloat64:    customFloat64{1.1, 1000.123, 999.0},
			CustomComplex128: customComplex128{100 + 1i, 10i, 10},
			CustomByte:       customByte{1, 2, 3},
			CustomStruct:     customStruct2{{"猿空"}, {"猿岳"}, {"猿河"}},
		},
	}
	tests := map[string]struct {
		prepare func(*Masker)
		input   Test
		want    Test
		isErr   bool
	}{
		"nil": {
			prepare: func(m *Masker) {},
			input: Test{
				Tag: Tag{
					ZeroGuard: "x",
				},
				NoTag: NoTag{
					ZeroGuard: "x",
				},
			},
			want: Test{
				Tag: Tag{
					ZeroGuard: "x",
				},
				NoTag: NoTag{
					ZeroGuard: "x",
				},
			},
		},
		"empty slice": {
			prepare: func(m *Masker) {},
			input: Test{
				Tag: Tag{
					String:           []string{},
					Int:              []int{},
					Uint:             []uint{},
					Float64:          []float64{},
					Complex128:       []complex128{},
					Byte:             []byte{},
					Struct:           []Struct1{},
					CustomString:     []string{},
					CustomInt:        []int{},
					CustomUint:       []uint{},
					CustomFloat64:    []float64{},
					CustomComplex128: []complex128{},
					CustomByte:       []byte{},
					CustomStruct:     customStruct{},
				},
				NoTag: NoTag{
					String:           []string{},
					Int:              []int{},
					Uint:             []uint{},
					Float64:          []float64{},
					Complex128:       []complex128{},
					Byte:             []byte{},
					Struct:           []Struct2{},
					CustomString:     []string{},
					CustomInt:        []int{},
					CustomUint:       []uint{},
					CustomFloat64:    []float64{},
					CustomComplex128: []complex128{},
					CustomByte:       []byte{},
					CustomStruct:     customStruct2{},
				},
			},
			want: Test{
				Tag: Tag{
					String:           []string{},
					Int:              []int{},
					Uint:             []uint{},
					Float64:          []float64{},
					Complex128:       []complex128{},
					Byte:             []byte{},
					Struct:           []Struct1{},
					CustomString:     []string{},
					CustomInt:        []int{},
					CustomUint:       []uint{},
					CustomFloat64:    []float64{},
					CustomComplex128: []complex128{},
					CustomByte:       []byte{},
					CustomStruct:     customStruct{},
				},
				NoTag: NoTag{
					String:           []string{},
					Int:              []int{},
					Uint:             []uint{},
					Float64:          []float64{},
					Complex128:       []complex128{},
					Byte:             []byte{},
					Struct:           []Struct2{},
					CustomString:     []string{},
					CustomInt:        []int{},
					CustomUint:       []uint{},
					CustomFloat64:    []float64{},
					CustomComplex128: []complex128{},
					CustomByte:       []byte{},
					CustomStruct:     customStruct2{},
				},
			},
		},
		"no masking functions": {
			prepare: func(m *Masker) {},
			input:   input,
			want:    input,
		},
		"register masking functions": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "test")
			},
			input: input,
			want: Test{
				Tag: Tag{
					String:           []string{"test", "test", "test"},
					Int:              []int{math.MaxInt, math.MaxInt, math.MaxInt},
					Uint:             []uint{math.MaxUint, math.MaxUint, math.MaxUint},
					Float64:          []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
					Complex128:       []complex128{100 + 1i, 10i, 10},
					Byte:             []byte{255, 255, 255},
					Struct:           []Struct1{{"test"}, {"test"}, {"test"}},
					CustomString:     []string{"test", "test", "test"},
					CustomInt:        []int{math.MaxInt, math.MaxInt, math.MaxInt},
					CustomUint:       []uint{math.MaxUint, math.MaxUint, math.MaxUint},
					CustomFloat64:    []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
					CustomComplex128: []complex128{100 + 1i, 10i, 10},
					CustomByte:       []byte{255, 255, 255},
					CustomStruct:     customStruct{{"test"}, {"test"}, {"test"}},
				},
				NoTag: input.NoTag,
			},
		},
		"register mask field name": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "field")
				m.RegisterMaskField("String", "field")
				m.RegisterMaskField("Int", "field")
				m.RegisterMaskField("Int8", "field")
				m.RegisterMaskField("Int16", "field")
				m.RegisterMaskField("Int32", "field")
				m.RegisterMaskField("Int64", "field")
				m.RegisterMaskField("Uint", "field")
				m.RegisterMaskField("Uint8", "field")
				m.RegisterMaskField("Uint16", "field")
				m.RegisterMaskField("Uint32", "field")
				m.RegisterMaskField("Uint64", "field")
				m.RegisterMaskField("Float32", "field")
				m.RegisterMaskField("Float64", "field")
				m.RegisterMaskField("Complex64", "field")
				m.RegisterMaskField("Complex128", "field")
				m.RegisterMaskField("Byte", "field")
				m.RegisterMaskField("Struct", "field")
				m.RegisterMaskField("CustomString", "field")
				m.RegisterMaskField("CustomInt", "field")
				m.RegisterMaskField("CustomInt8", "field")
				m.RegisterMaskField("CustomInt16", "field")
				m.RegisterMaskField("CustomInt32", "field")
				m.RegisterMaskField("CustomInt64", "field")
				m.RegisterMaskField("CustomUint", "field")
				m.RegisterMaskField("CustomUint8", "field")
				m.RegisterMaskField("CustomUint16", "field")
				m.RegisterMaskField("CustomUint32", "field")
				m.RegisterMaskField("CustomUint64", "field")
				m.RegisterMaskField("CustomFloat32", "field")
				m.RegisterMaskField("CustomFloat64", "field")
				m.RegisterMaskField("CustomComplex64", "field")
				m.RegisterMaskField("CustomComplex128", "field")
				m.RegisterMaskField("CustomByte", "field")
				m.RegisterMaskField("CustomStruct", "field")
			},
			input: input,
			want: Test{
				Tag: input.Tag,
				NoTag: NoTag{
					String:           []string{"test", "test", "test"},
					Int:              []int{math.MaxInt, math.MaxInt, math.MaxInt},
					Uint:             []uint{math.MaxUint, math.MaxUint, math.MaxUint},
					Float64:          []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
					Complex128:       []complex128{100 + 1i, 10i, 10},
					Byte:             []byte{255, 255, 255},
					Struct:           []Struct2{{"test"}, {"test"}, {"test"}},
					CustomString:     []string{"test", "test", "test"},
					CustomInt:        []int{math.MaxInt, math.MaxInt, math.MaxInt},
					CustomUint:       []uint{math.MaxUint, math.MaxUint, math.MaxUint},
					CustomFloat64:    []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
					CustomComplex128: []complex128{100 + 1i, 10i, 10},
					CustomByte:       []byte{255, 255, 255},
					CustomStruct:     customStruct2{{"test"}, {"test"}, {"test"}},
				},
			},
		},
	}

	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				m := NewMasker()
				tt.prepare(m)
				got, err := m.Mask(tt.input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_Map(t *testing.T) {
	type Key struct {
		Seq int
	}
	type customString map[string]string
	type customInt map[string]int
	type customUint map[string]uint
	type customFloat32 map[string]float32
	type customFloat64 map[string]float64
	type customComplex128 map[string]complex128
	type customByte map[string]byte
	type customIntKey map[int]string
	type customStructKey map[Key]string

	type Tag struct {
		String     map[string]string     `mask:"test"`
		Int        map[string]int        `mask:"test"`
		Uint       map[string]uint       `mask:"test"`
		Float32    map[string]float32    `mask:"test"`
		Float64    map[string]float64    `mask:"test"`
		Complex128 map[string]complex128 `mask:"test"`
		Byte       map[string]byte       `mask:"test"`
		IntKey     map[int]string        `mask:"test"`
		StructKey  map[Key]string        `mask:"test"`

		CustomString     customString     `mask:"test"`
		CustomInt        customInt        `mask:"test"`
		CustomUint       customUint       `mask:"test"`
		CustomFloat32    customFloat32    `mask:"test"`
		CustomFloat64    customFloat64    `mask:"test"`
		CustomComplex128 customComplex128 `mask:"test"`
		CustomByte       customByte       `mask:"test"`
		CustomIntKey     customIntKey     `mask:"test"`
		CustomStructKey  customStructKey  `mask:"test"`

		ZeroGuard string
	}
	type NoTag struct {
		String     map[string]string
		Int        map[string]int
		Uint       map[string]uint
		Float32    map[string]float32
		Float64    map[string]float64
		Complex128 map[string]complex128
		Byte       map[string]byte
		IntKey     map[int]string
		StructKey  map[Key]string

		CustomString     customString
		CustomInt        customInt
		CustomUint       customUint
		CustomFloat32    customFloat32
		CustomFloat64    customFloat64
		CustomComplex128 customComplex128
		CustomByte       customByte
		CustomIntKey     customIntKey
		CustomStructKey  customStructKey

		ZeroGuard string
	}
	type Test struct {
		Tag
		NoTag
	}

	input := Test{
		Tag: Tag{
			String:     map[string]string{"猿将": "大関", "猿谷": "小結", "猿桜": "三枚目"},
			Int:        map[string]int{"猿将": -1, "猿谷": 10, "猿桜": 100},
			Uint:       map[string]uint{"猿将": 1, "猿谷": 2, "猿桜": 3},
			Float32:    map[string]float32{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			Float64:    map[string]float64{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			Complex128: map[string]complex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
			Byte:       map[string]byte{"猿将": 1, "猿谷": 2, "猿桜": 3},
			IntKey:     map[int]string{1: "猿将", 2: "猿谷", 3: "猿桜"},
			StructKey:  map[Key]string{{1}: "猿将", {2}: "猿谷", {3}: "猿桜"},

			CustomString:     customString{"猿将": "大関", "猿谷": "小結", "猿桜": "三枚目"},
			CustomInt:        customInt{"猿将": -1, "猿谷": 10, "猿桜": 100},
			CustomUint:       customUint{"猿将": 1, "猿谷": 2, "猿桜": 3},
			CustomFloat32:    customFloat32{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			CustomFloat64:    customFloat64{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			CustomComplex128: customComplex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
			CustomByte:       customByte{"猿将": 1, "猿谷": 2, "猿桜": 3},
			CustomIntKey:     customIntKey{1: "猿将", 2: "猿谷", 3: "猿桜"},
			CustomStructKey:  customStructKey{{1}: "猿将", {2}: "猿谷", {3}: "猿桜"},
		},
		NoTag: NoTag{
			String:     map[string]string{"猿将": "大関", "猿谷": "小結", "猿桜": "三枚目"},
			Int:        map[string]int{"猿将": -1, "猿谷": 10, "猿桜": 100},
			Uint:       map[string]uint{"猿将": 1, "猿谷": 2, "猿桜": 3},
			Float32:    map[string]float32{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			Float64:    map[string]float64{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			Complex128: map[string]complex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
			Byte:       map[string]byte{"猿将": 1, "猿谷": 2, "猿桜": 3},
			IntKey:     map[int]string{1: "猿将", 2: "猿谷", 3: "猿桜"},
			StructKey:  map[Key]string{{1}: "猿将", {2}: "猿谷", {3}: "猿桜"},

			CustomString:     customString{"猿将": "大関", "猿谷": "小結", "猿桜": "三枚目"},
			CustomInt:        customInt{"猿将": -1, "猿谷": 10, "猿桜": 100},
			CustomUint:       customUint{"猿将": 1, "猿谷": 2, "猿桜": 3},
			CustomFloat32:    customFloat32{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			CustomFloat64:    customFloat64{"猿将": 1.1, "猿谷": 1000.123, "猿桜": 999.0},
			CustomComplex128: customComplex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
			CustomByte:       customByte{"猿将": 1, "猿谷": 2, "猿桜": 3},
			CustomIntKey:     customIntKey{1: "猿将", 2: "猿谷", 3: "猿桜"},
			CustomStructKey:  customStructKey{{1}: "猿将", {2}: "猿谷", {3}: "猿桜"},
		},
	}
	tests := map[string]struct {
		prepare func(*Masker)
		input   Test
		want    Test
		isErr   bool
	}{
		"nil": {
			prepare: func(*Masker) {},
			input: Test{
				Tag:   Tag{ZeroGuard: "x"},
				NoTag: NoTag{ZeroGuard: "x"},
			},
			want: Test{
				Tag:   Tag{ZeroGuard: "x"},
				NoTag: NoTag{ZeroGuard: "x"},
			},
		},
		"no masking functions": {
			prepare: func(m *Masker) {},
			input:   input,
			want:    input,
		},
		"register masking functions": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "test")
			},
			input: input,
			want: Test{
				Tag: Tag{
					String:     map[string]string{"猿将": "test", "猿谷": "test", "猿桜": "test"},
					Int:        map[string]int{"猿将": math.MaxInt, "猿谷": math.MaxInt, "猿桜": math.MaxInt},
					Uint:       map[string]uint{"猿将": math.MaxUint, "猿谷": math.MaxUint, "猿桜": math.MaxUint},
					Float32:    map[string]float32{"猿将": float32(math.Inf(0)), "猿谷": float32(math.Inf(0)), "猿桜": float32(math.Inf(0))},
					Float64:    map[string]float64{"猿将": math.MaxFloat64, "猿谷": math.MaxFloat64, "猿桜": math.MaxFloat64},
					Complex128: map[string]complex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
					Byte:       map[string]byte{"猿将": 255, "猿谷": 255, "猿桜": 255},
					IntKey:     map[int]string{1: "test", 2: "test", 3: "test"},
					StructKey:  map[Key]string{{1}: "test", {2}: "test", {3}: "test"},

					CustomString:     customString{"猿将": "test", "猿谷": "test", "猿桜": "test"},
					CustomInt:        customInt{"猿将": math.MaxInt, "猿谷": math.MaxInt, "猿桜": math.MaxInt},
					CustomUint:       customUint{"猿将": math.MaxUint, "猿谷": math.MaxUint, "猿桜": math.MaxUint},
					CustomFloat32:    customFloat32{"猿将": float32(math.Inf(0)), "猿谷": float32(math.Inf(0)), "猿桜": float32(math.Inf(0))},
					CustomFloat64:    customFloat64{"猿将": math.MaxFloat64, "猿谷": math.MaxFloat64, "猿桜": math.MaxFloat64},
					CustomComplex128: customComplex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
					CustomByte:       customByte{"猿将": 255, "猿谷": 255, "猿桜": 255},
					CustomIntKey:     customIntKey{1: "test", 2: "test", 3: "test"},
					CustomStructKey:  customStructKey{{1}: "test", {2}: "test", {3}: "test"},
				},
				NoTag: input.NoTag,
			},
		},
		"register mask field name": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "field")
				m.RegisterMaskField("String", "field")
				m.RegisterMaskField("Int", "field")
				m.RegisterMaskField("Int8", "field")
				m.RegisterMaskField("Int16", "field")
				m.RegisterMaskField("Int32", "field")
				m.RegisterMaskField("Int64", "field")
				m.RegisterMaskField("Uint", "field")
				m.RegisterMaskField("Uint8", "field")
				m.RegisterMaskField("Uint16", "field")
				m.RegisterMaskField("Uint32", "field")
				m.RegisterMaskField("Uint64", "field")
				m.RegisterMaskField("Float32", "field")
				m.RegisterMaskField("Float64", "field")
				m.RegisterMaskField("Complex64", "field")
				m.RegisterMaskField("Complex128", "field")
				m.RegisterMaskField("Byte", "field")
				m.RegisterMaskField("IntKey", "field")
				m.RegisterMaskField("StructKey", "field")

				m.RegisterMaskField("CustomString", "field")
				m.RegisterMaskField("CustomInt", "field")
				m.RegisterMaskField("CustomInt8", "field")
				m.RegisterMaskField("CustomInt16", "field")
				m.RegisterMaskField("CustomInt32", "field")
				m.RegisterMaskField("CustomInt64", "field")
				m.RegisterMaskField("CustomUint", "field")
				m.RegisterMaskField("CustomUint8", "field")
				m.RegisterMaskField("CustomUint16", "field")
				m.RegisterMaskField("CustomUint32", "field")
				m.RegisterMaskField("CustomUint64", "field")
				m.RegisterMaskField("CustomFloat32", "field")
				m.RegisterMaskField("CustomFloat64", "field")
				m.RegisterMaskField("CustomComplex64", "field")
				m.RegisterMaskField("CustomComplex128", "field")
				m.RegisterMaskField("CustomByte", "field")
				m.RegisterMaskField("CustomIntKey", "field")
				m.RegisterMaskField("CustomStructKey", "field")
				// map key
				m.RegisterMaskField("猿将", "field")
			},
			input: input,
			want: Test{
				Tag: input.Tag,
				NoTag: NoTag{
					String:     map[string]string{"猿将": "test", "猿谷": "test", "猿桜": "test"},
					Int:        map[string]int{"猿将": math.MaxInt, "猿谷": math.MaxInt, "猿桜": math.MaxInt},
					Uint:       map[string]uint{"猿将": math.MaxUint, "猿谷": math.MaxUint, "猿桜": math.MaxUint},
					Float32:    map[string]float32{"猿将": float32(math.Inf(0)), "猿谷": float32(math.Inf(0)), "猿桜": float32(math.Inf(0))},
					Float64:    map[string]float64{"猿将": math.MaxFloat64, "猿谷": math.MaxFloat64, "猿桜": math.MaxFloat64},
					Complex128: map[string]complex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
					Byte:       map[string]byte{"猿将": 255, "猿谷": 255, "猿桜": 255},
					IntKey:     map[int]string{1: "test", 2: "test", 3: "test"},
					StructKey:  map[Key]string{{1}: "test", {2}: "test", {3}: "test"},

					CustomString:     customString{"猿将": "test", "猿谷": "test", "猿桜": "test"},
					CustomInt:        customInt{"猿将": math.MaxInt, "猿谷": math.MaxInt, "猿桜": math.MaxInt},
					CustomUint:       customUint{"猿将": math.MaxUint, "猿谷": math.MaxUint, "猿桜": math.MaxUint},
					CustomFloat32:    customFloat32{"猿将": float32(math.Inf(0)), "猿谷": float32(math.Inf(0)), "猿桜": float32(math.Inf(0))},
					CustomFloat64:    customFloat64{"猿将": math.MaxFloat64, "猿谷": math.MaxFloat64, "猿桜": math.MaxFloat64},
					CustomComplex128: customComplex128{"猿将": 100 + 1i, "猿谷": 10i, "猿桜": 10},
					CustomByte:       customByte{"猿将": 255, "猿谷": 255, "猿桜": 255},
					CustomIntKey:     customIntKey{1: "test", 2: "test", 3: "test"},
					CustomStructKey:  customStructKey{{1}: "test", {2}: "test", {3}: "test"},
				},
			},
		},
	}

	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				m := NewMasker()
				tt.prepare(m)
				got, err := m.Mask(tt.input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_Pointer(t *testing.T) {
	type Struct1 struct {
		String string `mask:"test"`
	}
	type Struct2 struct {
		String string
	}
	type customMap map[string]string
	type customSlice []string
	type customArray [3]string
	type Tag struct {
		String      *string            `mask:"test"`
		Int         *int               `mask:"test"`
		Uint        *uint              `mask:"test"`
		Float64     *float64           `mask:"test"`
		Complex128  *complex128        `mask:"test"`
		Byte        *byte              `mask:"test"`
		Array       *[3]string         `mask:"test"`
		CustomArray *customArray       `mask:"test"`
		Slice       *[]string          `mask:"test"`
		CustomSlice *customSlice       `mask:"test"`
		Map         *map[string]string `mask:"test"`
		CustomMap   *customMap         `mask:"test"`
		Struct      *Struct1
		ZeroGuard   string
	}
	type NoTag struct {
		String      *string
		Int         *int
		Uint        *uint
		Float64     *float64
		Complex128  *complex128
		Byte        *byte
		Array       *[3]string
		CustomArray *customArray
		Slice       *[]string
		CustomSlice *customSlice
		Map         *map[string]string
		CustomMap   *customMap
		Struct      *Struct2
		ZeroGuard   string
	}
	type Test struct {
		Tag
		NoTag
	}
	input := Test{
		Tag: Tag{
			String:      convertStringPtr("龍谷"),
			Int:         convertIntPtr(123),
			Uint:        convertUintPtr(321),
			Float64:     convertFloat64Ptr(123.456),
			Complex128:  convertComplex128Ptr(123 + 456i),
			Byte:        convertBytePtr(2),
			Array:       &([3]string{"序ノ口", "序二段", "三枚目"}),
			CustomArray: &(customArray{"序ノ口", "序二段", "三枚目"}),
			Slice:       &[]string{"序ノ口", "序二段", "三枚目"},
			CustomSlice: &customSlice{"序ノ口", "序二段", "三枚目"},
			Map:         &map[string]string{"序ノ口": "石川", "序二段": "高橋", "三枚目": "猿河"},
			CustomMap:   &customMap{"序ノ口": "石川", "序二段": "高橋", "三枚目": "猿河"},
			Struct:      &Struct1{"稽古場"},
		},
		NoTag: NoTag{
			String:      convertStringPtr("龍谷"),
			Int:         convertIntPtr(123),
			Uint:        convertUintPtr(321),
			Float64:     convertFloat64Ptr(123.456),
			Complex128:  convertComplex128Ptr(123 + 456i),
			Byte:        convertBytePtr(2),
			Array:       &([3]string{"序ノ口", "序二段", "三枚目"}),
			CustomArray: &(customArray{"序ノ口", "序二段", "三枚目"}),
			Slice:       &[]string{"序ノ口", "序二段", "三枚目"},
			CustomSlice: &customSlice{"序ノ口", "序二段", "三枚目"},
			Map:         &map[string]string{"序ノ口": "石川", "序二段": "高橋", "三枚目": "猿河"},
			CustomMap:   &customMap{"序ノ口": "石川", "序二段": "高橋", "三枚目": "猿河"},
			Struct:      &Struct2{"稽古場"},
		},
	}

	tests := map[string]struct {
		prepare func(*Masker)
		input   Test
		want    Test
		isErr   bool
	}{
		"nil": {
			prepare: func(m *Masker) {},
			input: Test{
				Tag:   Tag{ZeroGuard: "x"},
				NoTag: NoTag{ZeroGuard: "x"},
			},
			want: Test{
				Tag:   Tag{ZeroGuard: "x"},
				NoTag: NoTag{ZeroGuard: "x"},
			},
		},
		"no masking functions": {
			prepare: func(m *Masker) {},
			input:   input,
			want:    input,
		},
		"register masking functions": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "test")
			},
			input: input,
			want: Test{
				Tag: Tag{
					String:      convertStringPtr("test"),
					Int:         convertIntPtr(math.MaxInt),
					Uint:        convertUintPtr(math.MaxUint),
					Float64:     convertFloat64Ptr(math.MaxFloat64),
					Complex128:  convertComplex128Ptr(123 + 456i),
					Byte:        convertBytePtr(255),
					Array:       &([3]string{"test", "test", "test"}),
					CustomArray: &(customArray{"test", "test", "test"}),
					Slice:       &[]string{"test", "test", "test"},
					CustomSlice: &customSlice{"test", "test", "test"},
					Map:         &map[string]string{"序ノ口": "test", "序二段": "test", "三枚目": "test"},
					CustomMap:   &customMap{"序ノ口": "test", "序二段": "test", "三枚目": "test"},
					Struct:      &Struct1{"test"},
				},
				NoTag: input.NoTag,
			},
		},
		"register mask field name": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "field")
				m.RegisterMaskField("String", "field")
				m.RegisterMaskField("Int", "field")
				m.RegisterMaskField("Uint", "field")
				m.RegisterMaskField("Float64", "field")
				m.RegisterMaskField("Complex128", "field")
				m.RegisterMaskField("Byte", "field")
				m.RegisterMaskField("Array", "field")
				m.RegisterMaskField("CustomArray", "field")
				m.RegisterMaskField("Slice", "field")
				m.RegisterMaskField("CustomSlice", "field")
				m.RegisterMaskField("Map", "field")
				m.RegisterMaskField("CustomMap", "field")
			},
			input: input,
			want: Test{
				Tag: input.Tag,
				NoTag: NoTag{
					String:      convertStringPtr("test"),
					Int:         convertIntPtr(math.MaxInt),
					Uint:        convertUintPtr(math.MaxUint),
					Float64:     convertFloat64Ptr(math.MaxFloat64),
					Complex128:  convertComplex128Ptr(123 + 456i),
					Byte:        convertBytePtr(255),
					Array:       &([3]string{"test", "test", "test"}),
					CustomArray: &(customArray{"test", "test", "test"}),
					Slice:       &[]string{"test", "test", "test"},
					CustomSlice: &customSlice{"test", "test", "test"},
					Map:         &map[string]string{"序ノ口": "test", "序二段": "test", "三枚目": "test"},
					CustomMap:   &customMap{"序ノ口": "test", "序二段": "test", "三枚目": "test"},
					Struct:      &Struct2{"test"},
				},
			},
		},
	}
	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				m := NewMasker()
				tt.prepare(m)
				got, err := m.Mask(tt.input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_Interface(t *testing.T) {
	type TestAny interface{}
	type Struct1 struct {
		String string `mask:"test"`
	}
	type Struct2 struct {
		String string
	}
	type Tag struct {
		String     TestAny `mask:"test"`
		Int        TestAny `mask:"test"`
		Uint       TestAny `mask:"test"`
		Float64    TestAny `mask:"test"`
		Complex128 TestAny `mask:"test"`
		Byte       TestAny `mask:"test"`
		Array      TestAny `mask:"test"`
		Slice      TestAny `mask:"test"`
		Map        TestAny `mask:"test"`
		Struct     TestAny
		Pointer    TestAny
		ZeroGuard  string
	}
	type NoTag struct {
		String     TestAny
		Int        TestAny
		Uint       TestAny
		Float64    TestAny
		Complex128 TestAny
		Byte       TestAny
		Array      TestAny
		Slice      TestAny
		Map        TestAny
		Struct     TestAny
		Pointer    TestAny
		ZeroGuard  string
	}
	type Test struct {
		Tag
		NoTag
	}
	input := Test{
		Tag: Tag{
			String:     "龍谷",
			Int:        123,
			Uint:       321,
			Float64:    123.456,
			Complex128: 123 + 456i,
			Byte:       2,
			Array:      [3]string{"序ノ口", "序二段", "三枚目"},
			Slice:      []string{"序ノ口", "序二段", "三枚目"},
			Map:        map[string]string{"序ノ口": "石川", "序二段": "高橋", "三枚目": "猿河"},
			Struct:     Struct1{"稽古場"},
			Pointer:    &Struct1{"稽古場"},
		},
		NoTag: NoTag{
			String:     "龍谷",
			Int:        123,
			Uint:       321,
			Float64:    123.456,
			Complex128: 123 + 456i,
			Byte:       2,
			Array:      [3]string{"序ノ口", "序二段", "三枚目"},
			Slice:      []string{"序ノ口", "序二段", "三枚目"},
			Map:        map[string]string{"序ノ口": "石川", "序二段": "高橋", "三枚目": "猿河"},
			Struct:     Struct2{"稽古場"},
			Pointer:    &Struct2{"稽古場"},
		},
	}

	tests := map[string]struct {
		prepare func(*Masker)
		input   Test
		want    Test
		isErr   bool
	}{
		"nil": {
			prepare: func(m *Masker) {},
			input: Test{
				Tag:   Tag{ZeroGuard: "x"},
				NoTag: NoTag{ZeroGuard: "x"},
			},
			want: Test{
				Tag:   Tag{ZeroGuard: "x"},
				NoTag: NoTag{ZeroGuard: "x"},
			},
		},
		"no masking functions": {
			prepare: func(m *Masker) {},
			input:   input,
			want:    input,
		},
		"register masking functions": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "test")
			},
			input: input,
			want: Test{
				Tag: Tag{
					String:     "test",
					Int:        math.MaxInt,
					Uint:       math.MaxInt,
					Float64:    math.MaxFloat64,
					Complex128: 123 + 456i,
					Byte:       math.MaxInt,
					Array:      [3]string{"test", "test", "test"},
					Slice:      []string{"test", "test", "test"},
					Map:        map[string]string{"序ノ口": "test", "序二段": "test", "三枚目": "test"},
					Struct:     Struct1{"test"},
					Pointer:    &Struct1{"test"},
				},
				NoTag: input.NoTag,
			},
		},
		"register mask field name": {
			prepare: func(m *Masker) {
				registerTestMaskFunc(m, "field")
				m.RegisterMaskField("String", "field")
				m.RegisterMaskField("Int", "field")
				m.RegisterMaskField("Uint", "field")
				m.RegisterMaskField("Float64", "field")
				m.RegisterMaskField("Complex128", "field")
				m.RegisterMaskField("Byte", "field")
				m.RegisterMaskField("Array", "field")
				m.RegisterMaskField("Slice", "field")
				m.RegisterMaskField("Map", "field")
			},
			input: input,
			want: Test{
				Tag: input.Tag,
				NoTag: NoTag{
					String:     "test",
					Int:        math.MaxInt,
					Uint:       math.MaxInt,
					Float64:    math.MaxFloat64,
					Complex128: 123 + 456i,
					Byte:       math.MaxInt,
					Array:      [3]string{"test", "test", "test"},
					Slice:      []string{"test", "test", "test"},
					Map:        map[string]string{"序ノ口": "test", "序二段": "test", "三枚目": "test"},
					Struct:     Struct2{"test"},
					Pointer:    &Struct2{"test"},
				},
			},
		},
	}
	for name, tt := range tests {
		for _, cache := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s - cache enable=%t", name, cache), func(t *testing.T) {
				m := NewMasker()
				tt.prepare(m)
				got, err := m.Mask(tt.input)
				if tt.isErr {
					if err == nil {
						t.Error("want an error to occur")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

func TestMask_SameStruct(t *testing.T) {
	type sameStructNameTest struct {
		Usagi string
	}
	createSameStruct := func(value int) any {
		type sameStructNameTest struct {
			Usagi int
		}
		return sameStructNameTest{value}
	}

	t.Run(defaultTestCase("same struct name"), func(t *testing.T) {
		defer cleanup(t)
		{
			input := sameStructNameTest{"Rabbit"}
			got, err := Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}
		{
			input := createSameStruct(2)
			got, err := Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}
	})
	t.Run(newMaskerTestCase("same struct name"), func(t *testing.T) {
		m := newMasker()
		{
			input := sameStructNameTest{"Rabbit"}
			got, err := m.Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}

		{
			input := createSameStruct(2)
			got, err := m.Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}
	})
}

func TestMask_SameAnonynousStruct(t *testing.T) {
	t.Run(defaultTestCase("same anonymous struct name"), func(t *testing.T) {
		defer cleanup(t)
		{
			input := struct {
				Usagi string
			}{
				Usagi: "Rabbit",
			}
			got, err := Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}
		{
			input := struct {
				A int
			}{
				A: 2,
			}
			got, err := Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}
	})
	t.Run(newMaskerTestCase("same anonymous struct name"), func(t *testing.T) {
		m := newMasker()
		{
			input := struct {
				Usagi string
			}{
				Usagi: "Rabbit",
			}
			got, err := m.Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}

		{
			input := struct {
				A int
			}{
				A: 2,
			}
			got, err := m.Mask(input)
			assert.Nil(t, err)
			if diff := cmp.Diff(input, got); diff != "" {
				t.Error(diff)
			}
		}
	})
}

func registerTestMaskFunc(m *Masker, tag string) {
	m.RegisterMaskStringFunc(tag, func(arg, value string) (string, error) {
		return "test", nil
	})
	m.RegisterMaskIntFunc(tag, func(arg string, value int) (int, error) {
		return math.MaxInt, nil
	})
	m.RegisterMaskUintFunc(tag, func(arg string, value uint) (uint, error) {
		return math.MaxUint, nil
	})
	m.RegisterMaskFloat64Func(tag, func(arg string, value float64) (float64, error) {
		return math.MaxFloat64, nil
	})
}

func TestSetTagName(t *testing.T) {
	t.Run("change a tag name", func(t *testing.T) {
		m := newMasker()
		m.SetTagName("fake")

		input := struct {
			SM string `mask:"filled4"`
			SF string `fake:"filled4"`
		}{
			SM: "Hello World",
			SF: "Hello World",
		}
		want := struct {
			SM string `mask:"filled4"`
			SF string `fake:"filled4"`
		}{
			SM: "Hello World",
			SF: "****",
		}
		got, _ := m.Mask(input)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("change a empty tag name", func(t *testing.T) {
		m := newMasker()
		m.SetTagName("")

		input := struct {
			SM string `mask:"filled4"`
			SF string `fake:"filled4"`
		}{
			SM: "Hello World",
			SF: "Hello World",
		}
		want := struct {
			SM string `mask:"filled4"`
			SF string `fake:"filled4"`
		}{
			SM: "****",
			SF: "Hello World",
		}
		got, _ := m.Mask(input)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})
}

func TestSetMaskChar(t *testing.T) {
	t.Run("change a mask character", func(t *testing.T) {
		defer cleanup(t)
		SetMaskChar("-")

		input := struct {
			S string `mask:"filled4"`
		}{
			S: "Hello World",
		}
		want := struct {
			S string `mask:"filled4"`
		}{
			S: "----",
		}
		got, _ := Mask(input)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("change a empty mask character", func(t *testing.T) {
		defer cleanup(t)
		SetMaskChar("")

		input := struct {
			S string `mask:"filled4"`
		}{
			S: "Hello World",
		}
		want := struct {
			S string `mask:"filled4"`
		}{
			S: "",
		}
		got, _ := Mask(input)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})
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
	type stringCustomSlice []string
	type customStringSliceTest struct {
		Usagi stringCustomSlice `mask:"filled"`
	}
	type stringSlicePtrTest struct {
		Usagi *[]string `mask:"filled"`
	}
	type stringToStringMapTest struct {
		Usagi map[string]string `mask:"filled"`
	}
	type stringToStringMap map[string]string
	type customStringToStringMapTest struct {
		Usagi stringToStringMap `mask:"filled"`
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
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！"},
			want:  &stringTest{Usagi: "****"},
		},
		"zero string fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
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
		"custom string slice fields": {
			input: &customStringSliceTest{Usagi: stringCustomSlice{"ハァ？", "ウラ", "フゥン"}},
			want:  &customStringSliceTest{Usagi: stringCustomSlice{"***", "**", "***"}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"nil custom string slice fields": {
			input: &customStringSliceTest{},
			want:  &customStringSliceTest{Usagi: (stringCustomSlice)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: &([]string{"ハァ？", "ウラ", "フゥン"})},
			want:  &stringSlicePtrTest{Usagi: &([]string{"***", "**", "***"})},
		},
		"nil string slice ptr fields": {
			input: &stringSlicePtrTest{},
			want:  &stringSlicePtrTest{Usagi: (*[]string)(nil)},
		},
		"string to string map fields": {
			input: &stringToStringMapTest{Usagi: map[string]string{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want:  &stringToStringMapTest{Usagi: map[string]string{"うさぎ": "***", "うさぎ2": "**", "うさぎ3": "***"}},
		},
		"custom string to string map fields": {
			input: &customStringToStringMapTest{Usagi: stringToStringMap{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want:  &customStringToStringMapTest{Usagi: stringToStringMap{"うさぎ": "***", "うさぎ2": "**", "うさぎ3": "***"}},
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
		t.Run(defaultTestCase(name), func(t *testing.T) {
			defer cleanup(t)
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
		t.Run(newMaskerTestCase(name), func(t *testing.T) {
			m := newMasker()
			got, err := m.Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskFixed(t *testing.T) {
	type stringTest struct {
		Usagi string `mask:"fixed"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！！！"},
			want:  &stringTest{Usagi: "********"},
		},
		"zero string fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
		},
	}

	for name, tt := range tests {
		t.Run(defaultTestCase(name), func(t *testing.T) {
			defer cleanup(t)
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
		t.Run(newMaskerTestCase(name), func(t *testing.T) {
			m := newMasker()
			got, err := m.Mask(tt.input)
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
	type stringCustomSlice []string
	type stringCustomSliceTest struct {
		Usagi stringCustomSlice `mask:"hash"`
	}
	type stringArrayTest struct {
		Usagi [3]string `mask:"hash"`
	}
	type stringArray [3]string
	type stringCustomArrayTest struct {
		Usagi stringArray `mask:"hash"`
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
	type stringMap map[string]string
	type structToCustomStringMapTest struct {
		Usagi stringMap `mask:"hash"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string fields": {
			input: &stringTest{Usagi: "ヤハッ！"},
			want:  &stringTest{Usagi: "a6ab5728db57954641b2e155adc61f2cbdfc7063"},
		},
		"zero string fields": {
			input: &stringTest{},
			want:  &stringTest{Usagi: ""},
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
		"string custom slice fields": {
			input: &stringCustomSliceTest{Usagi: stringCustomSlice{"ハァ？", "ウラ", "フゥン"}},
			want: &stringCustomSliceTest{Usagi: stringCustomSlice{
				"48a8b33f36a35631f584844686adaba89a6f156a",
				"ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"string array fields": {
			input: &stringArrayTest{Usagi: [3]string{"ハァ？", "ウラ", "フゥン"}},
			want: &stringArrayTest{Usagi: [3]string{
				"48a8b33f36a35631f584844686adaba89a6f156a",
				"ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"custom string array fields": {
			input: &stringCustomArrayTest{Usagi: stringArray{"ハァ？", "ウラ", "フゥン"}},
			want: &stringCustomArrayTest{Usagi: stringArray{
				"48a8b33f36a35631f584844686adaba89a6f156a",
				"ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"nil custom string slice fields": {
			input: &stringCustomSliceTest{},
			want:  &stringCustomSliceTest{Usagi: (stringCustomSlice)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: &([]string{"ハァ？", "ウラ", "フゥン"})},
			want: &stringSlicePtrTest{Usagi: &([]string{
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
			want: &structToStringMapTest{Usagi: map[stringTest]string{
				{Usagi: "ヤハッ！"}:   "48a8b33f36a35631f584844686adaba89a6f156a",
				{Usagi: "ヤハッ！！"}:  "ecef3e43f07f7150c089e99d5e1041259b1189d5",
				{Usagi: "ヤハッ！！！"}: "17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
		"struct to custom string map fields": {
			input: &structToCustomStringMapTest{Usagi: stringMap{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want: &structToCustomStringMapTest{Usagi: stringMap{
				"うさぎ":  "48a8b33f36a35631f584844686adaba89a6f156a",
				"うさぎ2": "ecef3e43f07f7150c089e99d5e1041259b1189d5",
				"うさぎ3": "17fa078ad3f2c34c17ee58b9119963548ddcf1ef",
			}},
		},
	}

	for name, tt := range tests {
		t.Run(defaultTestCase(name), func(t *testing.T) {
			defer cleanup(t)
			got, err := Mask(tt.input)
			assert.Nil(t, err)

			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
		t.Run(newMaskerTestCase(name), func(t *testing.T) {
			m := newMasker()
			got, err := m.Mask(tt.input)
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
	type int16Test struct {
		Usagi int32 `mask:"random1000"`
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
	type customIntSlice []int
	type customIntSliceTest struct {
		Usagi customIntSlice `mask:"random1000"`
	}
	type int32SliceTest struct {
		Usagi []int32 `mask:"random1000"`
	}
	type int64SliceTest struct {
		Usagi []int64 `mask:"random1000"`
	}
	type intArrayTest struct {
		Usagi [2]int `mask:"random1000"`
	}
	type customIntArray [2]int
	type customIntArrayTest struct {
		Usagi customIntArray `mask:"random1000"`
	}
	type int32ArrayTest struct {
		Usagi [2]int32 `mask:"random1000"`
	}
	type int64ArrayTest struct {
		Usagi [2]int64 `mask:"random1000"`
	}
	type intSlicePtrTest struct {
		Usagi *[]int `mask:"random1000"`
	}
	type float32Test struct {
		Usagi float32 `mask:"random100000.4"`
	}
	type float64Test struct {
		Usagi float64 `mask:"random100000.4"`
	}
	type float64PtrTest struct {
		Usagi *float64 `mask:"random100000.4"`
	}
	type float32SliceTest struct {
		Usagi []float32 `mask:"random100000.4"`
	}
	type float64SliceTest struct {
		Usagi []float64 `mask:"random100000.4"`
	}
	type float32ArrayTest struct {
		Usagi [3]float32 `mask:"random100000.4"`
	}
	type float64ArrayTest struct {
		Usagi [3]float64 `mask:"random100000.4"`
	}
	type float64SlicePtrTest struct {
		Usagi *[]float64 `mask:"random100000.4"`
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
	type stringToInt map[string]int
	type customStringToMapTest struct {
		Usagi stringToInt `mask:"random1000"`
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"int fields": {
			input: &intTest{Usagi: 20190122},
			want:  &intTest{Usagi: 829},
		},
		"int16 fields": {
			input: &int16Test{Usagi: 2019},
			want:  &int16Test{Usagi: 829},
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
			input: &customIntSliceTest{Usagi: customIntSlice{20190122, 20200501, 20200501}},
			want:  &customIntSliceTest{Usagi: customIntSlice{829, 830, 400}},
		},
		"int32 slice fields": {
			input: &int32SliceTest{Usagi: []int32{20190122, 20200501, 20200501}},
			want:  &int32SliceTest{Usagi: []int32{829, 830, 400}},
		},
		"int64 slice fields": {
			input: &int64SliceTest{Usagi: []int64{20190122, 20200501, 20200501}},
			want:  &int64SliceTest{Usagi: []int64{829, 830, 400}},
		},
		"int array fields": {
			input: &intArrayTest{Usagi: [2]int{20190122, 20200501}},
			want:  &intArrayTest{Usagi: [2]int{829, 830}},
		},
		"int32 array fields": {
			input: &int32ArrayTest{Usagi: [2]int32{20190122, 20200501}},
			want:  &int32ArrayTest{Usagi: [2]int32{829, 830}},
		},
		"custom int32 array fields": {
			input: &customIntArrayTest{Usagi: customIntArray{20190122, 20200501}},
			want:  &customIntArrayTest{Usagi: customIntArray{829, 830}},
		},
		"int64 array fields": {
			input: &int64ArrayTest{Usagi: [2]int64{20190122, 20200501}},
			want:  &int64ArrayTest{Usagi: [2]int64{829, 830}},
		},
		"nil int slice fields": {
			input: &intSliceTest{},
			want:  &intSliceTest{Usagi: ([]int)(nil)},
		},
		"int slice ptr fields": {
			input: &intSlicePtrTest{Usagi: &([]int{20190122, 20200501, 20200501})},
			want:  &intSlicePtrTest{Usagi: &([]int{829, 830, 400})},
		},
		"nil int slice ptr fields": {
			input: &intSlicePtrTest{},
			want:  &intSlicePtrTest{Usagi: (*[]int)(nil)},
		},
		"float32 fields": {
			input: &float32Test{Usagi: 20190122},
			want:  &float32Test{Usagi: 96011.8989},
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
		"float32 array fields": {
			input: &float32ArrayTest{Usagi: [3]float32{20190122, 20200501, 20200501}},
			want:  &float32ArrayTest{Usagi: [3]float32{96011.8989, 90863.3149, 32310.0201}},
		},
		"float64 array fields": {
			input: &float64ArrayTest{Usagi: [3]float64{20190122, 20200501, 20200501}},
			want:  &float64ArrayTest{Usagi: [3]float64{96011.8989, 90863.3149, 32310.0201}},
		},
		"nil float64 slice fields": {
			input: &float64SliceTest{},
			want:  &float64SliceTest{Usagi: ([]float64)(nil)},
		},
		"float64 slice ptr fields": {
			input: &float64SlicePtrTest{Usagi: &([]float64{20190122, 20200501, 20200501})},
			want:  &float64SlicePtrTest{Usagi: &([]float64{96011.8989, 90863.3149, 32310.0201})},
		},
		"nil float64 slice ptr fields": {
			input: &float64SlicePtrTest{},
			want:  &float64SlicePtrTest{Usagi: (*[]float64)(nil)},
		},
		"string to int map fields": {
			input: &stringToIntTest{Usagi: map[string]int{"うさぎ": 20190122}},
			want:  &stringToIntTest{Usagi: map[string]int{"うさぎ": 829}},
		},
		"string to int32 map fields": {
			input: &stringToInt32Test{Usagi: map[string]int32{"うさぎ": 20190122}},
			want:  &stringToInt32Test{Usagi: map[string]int32{"うさぎ": 829}},
		},
		"string to int64 map fields": {
			input: &stringToInt64Test{Usagi: map[string]int64{"うさぎ": 20190122}},
			want:  &stringToInt64Test{Usagi: map[string]int64{"うさぎ": 829}},
		},
		"string to custom int map fields": {
			input: &customStringToMapTest{Usagi: stringToInt{"うさぎ": 20190122}},
			want:  &customStringToMapTest{Usagi: stringToInt{"うさぎ": 829}},
		},
	}

	for name, tt := range tests {
		t.Run(defaultTestCase(name), func(t *testing.T) {
			defer cleanup(t)
			rand.Seed(rand.NewSource(1).Int63())
			got, err := Mask(tt.input)
			if assert.NoError(t, err) {
				if diff := cmp.Diff(tt.want, got, cmpopts.SortMaps(func(i, j string) bool { return i < j })); diff != "" {
					t.Error(diff)
				}
			}
		})

		t.Run(newMaskerTestCase(name), func(t *testing.T) {
			rand.Seed(rand.NewSource(1).Int63())
			m := newMasker()
			got, err := m.Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMaskZero(t *testing.T) {
	type stringTest struct {
		Usagi string `mask:"zero"`
	}
	type stringPtrTest struct {
		Usagi *string `mask:"zero"`
	}
	type stringSliceTest struct {
		Usagi []string `mask:"zero"`
	}
	type stringArrayTest struct {
		Usagi [3]string `mask:"zero"`
	}
	type stringSlicePtrTest struct {
		Usagi *[]string `mask:"zero"`
	}
	type intTest struct {
		Usagi int `mask:"zero"`
	}
	type uintTest struct {
		Usagi uint `mask:"zero"`
	}
	type float64Test struct {
		Usagi float64 `mask:"zero"`
	}
	type boolTest struct {
		Usagi bool `mask:"zero"`
	}
	type mapStringToStringTest struct {
		Usagi map[string]string `mask:"zero"`
	}
	type structTest struct {
		StringTest stringTest `mask:"zero"`
	}
	type stringMap map[string]string
	type customStringMapTest struct {
		Usagi stringMap `mask:"zero"`
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
		"string array fields": {
			input: &stringArrayTest{Usagi: [3]string{"ハァ？", "ウラ", "フゥン"}},
			want:  &stringArrayTest{Usagi: [3]string{}},
		},
		"nil string slice fields": {
			input: &stringSliceTest{},
			want:  &stringSliceTest{Usagi: ([]string)(nil)},
		},
		"string slice ptr fields": {
			input: &stringSlicePtrTest{Usagi: &([]string{"ハァ？", "ウラ", "フゥン"})},
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
		"uint fields": {
			input: &uintTest{Usagi: 20190122},
			want:  &uintTest{Usagi: 0},
		},
		"zero uint fields": {
			input: &uintTest{},
			want:  &uintTest{Usagi: 0},
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
		"custom map string to string fields": {
			input: &customStringMapTest{Usagi: stringMap{"うさぎ": "ハァ？", "うさぎ2": "ウラ", "うさぎ3": "フゥン"}},
			want:  &customStringMapTest{},
		},
		"custom nil map string to string fields": {
			input: &customStringMapTest{},
			want:  &customStringMapTest{},
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
		t.Run(defaultTestCase(name), func(t *testing.T) {
			defer cleanup(t)
			rand.Seed(rand.NewSource(1).Int63())
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})

		t.Run(newMaskerTestCase(name), func(t *testing.T) {
			rand.Seed(rand.NewSource(1).Int63())
			m := newMasker()
			got, err := m.Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got, allowUnexported(tt.input)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCyclicGraph(t *testing.T) {
	const elCnt = 9
	type structTest struct {
		Usagi              string `mask:"filled"`
		Slice              []interface{}
		SlicePtr           *[]interface{}
		Array              [elCnt]interface{}
		ArrayPtr           *[elCnt]interface{}
		StructPtr          *structTest
		Map                map[string]interface{}
		MapPtr             *map[string]interface{}
		Interface          interface{}
		InterfaceStructPtr interface{}
	}

	createDataSet := func(val string) *structTest {
		cyclicSlice := make([]interface{}, elCnt)
		cyclicArray := [elCnt]interface{}{}

		cyclicStruct := structTest{
			Usagi: val,
		}

		cyclicMap := map[string]interface{}{
			"Usagi":     "ヤハッ！",
			"StructPtr": &cyclicStruct,
			"ArrayPtr":  &cyclicArray,
			"SlicePtr":  &cyclicSlice,
			"Array":     cyclicArray,
			// cmp.Diff does not support cycling this way
			//"Slice": cyclicSlice,
		}
		cyclicMap["MapPtr"] = &cyclicMap
		cyclicMap["Map"] = cyclicMap

		cyclicSlice[0] = "ヤハッ！"
		cyclicSlice[1] = &cyclicMap
		cyclicSlice[2] = &cyclicStruct
		cyclicSlice[3] = &cyclicArray
		cyclicSlice[4] = &cyclicSlice
		cyclicSlice[5] = cyclicMap
		cyclicSlice[7] = cyclicArray
		// cmp.Diff does not support cycling this way
		//cyclicSlice[8] = cyclicSlice

		cyclicArray[0] = "ヤハッ！"
		cyclicArray[1] = &cyclicMap
		cyclicArray[2] = &cyclicStruct
		cyclicArray[3] = &cyclicArray
		cyclicArray[4] = &cyclicSlice
		cyclicArray[5] = cyclicMap
		cyclicArray[7] = cyclicArray
		// cmp.Diff does not support cycling this way
		//cyclicArray[8] = cyclicSlice

		cyclicStruct.MapPtr = &cyclicMap
		cyclicStruct.StructPtr = &cyclicStruct
		cyclicStruct.ArrayPtr = &cyclicArray
		cyclicStruct.SlicePtr = &cyclicSlice
		cyclicStruct.Map = cyclicMap
		cyclicStruct.Array = cyclicArray
		// cmp.Diff does not support cycling this way
		//cyclicStruct.Slice = cyclicSlice
		cyclicStruct.Interface = cyclicStruct
		cyclicStruct.InterfaceStructPtr = &cyclicStruct

		// cmp.Diff does not support cycling this way
		//cyclicArray[6] = cyclicStruct
		//cyclicSlice[6] = cyclicStruct

		return &cyclicStruct
	}

	ds := createDataSet("ヤハッ！")
	expected := createDataSet("****")

	tests := map[string]struct {
		input any
		want  any
	}{
		"slice": {
			input: ds.Slice,
			want:  expected.Slice,
		},
		"array": {
			input: ds.Array,
			want:  expected.Array,
		},
		"struct": {
			input: *ds,
			want:  *expected,
		},
		"struct ptr": {
			input: ds,
			want:  expected,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, isCached := range []bool{false, true} {
				name := "uncached"
				if isCached {
					name = "cached"
				}
				t.Run(name, func(t *testing.T) {
					m := newMasker()
					m.Cache(isCached)
					got, err := m.Mask(tt.input)
					assert.Nil(t, err)
					if diff := cmp.Diff(tt.want, got); diff != "" {
						t.Error(diff)
					}
				})
			}
		})
	}
}

func TestAnyMaskFunc(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		m := newMasker()
		m.RegisterMaskAnyFunc("test", func(arg string, value any) (any, error) {
			return "白鳳", nil
		})

		got, err := m.String("test", "朝青龍")
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff("白鳳", got); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("Int", func(t *testing.T) {
		m := newMasker()
		m.RegisterMaskAnyFunc("test", func(arg string, value any) (any, error) {
			return 33, nil
		})

		got, err := m.Int("test", 11)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(33, got); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("Uint", func(t *testing.T) {
		m := newMasker()
		m.RegisterMaskAnyFunc("test", func(arg string, value any) (any, error) {
			return uint(44), nil
		})

		got, err := m.Uint("test", 11)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(uint(44), got); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("Float64", func(t *testing.T) {
		m := newMasker()
		m.RegisterMaskAnyFunc("test", func(arg string, value any) (any, error) {
			return 123.45, nil
		})

		got, err := m.Float64("test", 12.3)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(123.45, got); diff != "" {
			t.Error(diff)
		}
	})
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
	case reflect.Ptr, reflect.Slice:
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
func convertIntPtr(i int) *int {
	return &i
}
func convertUintPtr(i uint) *uint {
	return &i
}
func convertBytePtr(v byte) *byte {
	return &v
}
func convertFloat64Ptr(f float64) *float64 {
	return &f
}
func convertComplex128Ptr(c complex128) *complex128 {
	return &c
}
func convertBoolPtr(v bool) *bool {
	return &v
}
func convertAnyPtr(v any) *any {
	return &v
}

func defaultTestCase(name string) string {
	return "default Masker:" + name
}
func newMaskerTestCase(name string) string {
	return "newMasker:" + name
}

func cleanup(t *testing.T) {
	t.Helper()
	defaultMasker.cb = sharedCircuitBreaker{}
	SetMaskChar(maskChar)
}

func newMasker() *Masker {
	m := NewMasker()
	m.RegisterMaskStringFunc(MaskTypeFilled, m.MaskFilledString)
	m.RegisterMaskStringFunc(MaskTypeFixed, m.MaskFixedString)
	m.RegisterMaskStringFunc(MaskTypeHash, m.MaskHashString)
	m.RegisterMaskIntFunc(MaskTypeRandom, m.MaskRandomInt)
	m.RegisterMaskFloat64Func(MaskTypeRandom, m.MaskRandomFloat64)
	m.RegisterMaskAnyFunc(MaskTypeZero, m.MaskZero)
	return m
}
