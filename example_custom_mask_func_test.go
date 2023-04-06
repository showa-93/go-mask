package mask_test

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	mask "github.com/showa-93/go-mask"
)

func init() {
	maskTypeRegExp := "regexp"
	mask.RegisterMaskStringFunc(maskTypeRegExp, MaskRegExp)
}

// MaskRegExp is sample to add a custom mask function
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
		sb.WriteString(strings.Repeat(mask.MaskChar(), utf8.RuneCountInString(value[indexes[2]:indexes[3]])))
		sb.WriteString(value[indexes[3]:])
		return sb.String(), nil
	}

	return value, nil
}

func Example_customMaskFunc() {
	mask.SetMaskChar("■")
	type Hachiware struct {
		Message string `mask:"regexp(最高)."`
	}

	input := Hachiware{Message: "これって…最高じゃん"}
	got, _ := mask.Mask(input)
	fmt.Printf("\"%s\", Hachiware says\n", got.(Hachiware).Message)

	// Output:
	// "これって…■■じゃん", Hachiware says
}
