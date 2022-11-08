package maskgo_test

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/showa-93/maskgo"
)

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
		sb.WriteString(strings.Repeat(maskgo.MaskChar(), utf8.RuneCountInString(value[indexes[2]:indexes[3]])))
		sb.WriteString(value[indexes[3]:])
		return sb.String(), nil
	}

	return value, nil
}

func Example_wholeFileExample() {
	maskTypeRegExp := "regexp"
	maskgo.RegisterMaskStringFunc(maskTypeRegExp, MaskRegExp)

	type RegExpTest struct {
		Usagi string `mask:"regexp(ヤハ)*"`
	}

	input := RegExpTest{Usagi: "ヤハッ！"}
	got, _ := maskgo.Mask(input)
	fmt.Printf("Usagi %s\n", got.(RegExpTest).Usagi)

	// Output:
	// Usagi **ッ！
}
