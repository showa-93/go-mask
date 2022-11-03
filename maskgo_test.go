package maskgo

import (
	"testing"

	"github.com/ggwhite/go-masker"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestMask(t *testing.T) {
	type stringTest struct {
		Usagi string
	}

	tests := map[string]struct {
		input any
		want  any
	}{
		"string fields": {
			input: &stringTest{Usagi: "ウラッ"},
			want:  &stringTest{Usagi: "ウラッ"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Mask(tt.input)
			assert.Nil(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

type benchStruct2 struct {
	Case1 string `mask:"name"`
	Case2 int
	Case3 bool
	Case4 []string `mask:"name"`
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

func convertStringPtr(s string) *string {
	return &s
}
