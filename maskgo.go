package maskgo

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/goccy/go-reflect"
)

const tagName = "mask"

const (
	MaskTypeAll     = "all"
	maskTypeUnknown = "unknown"
)

type storeStruct struct {
	rv           reflect.Value
	structFields []reflect.StructField
}

type maskStringFunc func(arg, value string) (string, error)

var (
	typeToStruct      sync.Map
	maskChar                                    = "*"
	maskStringFuncMap map[string]maskStringFunc = map[string]maskStringFunc{
		MaskTypeAll:     maskAllString,
		maskTypeUnknown: func(_, value string) (string, error) { return value, nil },
		"name":          maskNameString,
	}
)

func RegisterMaskFunc(maskType string, f maskStringFunc) {
	maskStringFuncMap[maskType] = f
}

func MaskString(tag, value string) (string, error) {
	var (
		ok             bool
		arg            string
		maskStringFunc maskStringFunc
	)
	if tag != "" {
		for mt, f := range maskStringFuncMap {
			if strings.HasPrefix(tag, mt) {
				ok = true
				maskStringFunc = f
				arg = tag[len(mt):]
				break
			}
		}
	}
	if !ok {
		maskStringFunc = maskStringFuncMap[maskTypeUnknown]
	}
	return maskStringFunc(arg, value)
}

func Mask(target any) (any, error) {
	if target == nil {
		return target, nil
	}
	rv, err := mask(reflect.ValueOf(target), "")
	if err != nil {
		return nil, err
	}
	return rv.Interface(), nil
}

func mask(rv reflect.Value, tag string) (reflect.Value, error) {
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return rv, nil
		}
		rv2 := reflect.ValueOf(rv.Interface())
		rv3, err := mask(rv2.Elem(), tag)
		if err != nil {
			return reflect.Value{}, err
		}
		rv2.Elem().Set(rv3)
		return rv2, nil
	case reflect.Struct:
		if rv.IsZero() {
			return rv, nil
		}

		var ss storeStruct
		if storeValue, ok := typeToStruct.Load(rt.Name()); ok {
			ss = storeValue.(storeStruct)
		} else {
			ss.rv = reflect.New(rt).Elem()
			ss.structFields = make([]reflect.StructField, rv.NumField())
			for i := 0; i < rv.NumField(); i++ {
				ss.structFields[i] = rt.Field(i)
			}

			typeToStruct.Store(rt.Name(), ss)
		}

		for i := 0; i < rv.NumField(); i++ {
			if ss.structFields[i].PkgPath != "" {
				continue
			}
			vTag := ss.structFields[i].Tag.Get(tagName)
			rvf, err := mask(rv.Field(i), vTag)
			if err != nil {
				return reflect.Value{}, err
			}
			if rvf.IsValid() {
				ss.rv.Field(i).Set(rvf)
			}
		}

		return ss.rv, nil
	case reflect.Slice:
		if rv.IsZero() {
			return rv, nil
		}

		rv2 := reflect.MakeSlice(rt, rv.Len(), rv.Len())
		if rt.Elem().Kind() == reflect.String {
			for i, str := range rv.Interface().([]string) {
				rvf, err := MaskString(tag, str)
				if err != nil {
					return reflect.Value{}, err
				}
				rv2.Index(i).SetString(rvf)
			}
		} else {
			for i := 0; i < rv.Len(); i++ {
				rf, err := mask(rv.Index(i), tag)
				if err != nil {
					return reflect.Value{}, err
				}
				rv2.Index(i).Set(rf)
			}
		}

		return rv2, nil
	case reflect.String:
		sp, err := MaskString(tag, rv.String())
		if err != nil {
			return reflect.Value{}, err
		}
		rv2 := reflect.ValueOf(&sp)
		return rv2.Elem(), nil
	default:
		return rv, nil
	}
}

func maskAllString(arg, value string) (string, error) {
	return strings.Repeat(maskChar, utf8.RuneCountInString(value)), nil
}

func maskNameString(arg, value string) (string, error) {
	l := len([]rune(value))

	if l == 0 {
		return "", nil
	}

	if strs := strings.Split(value, " "); len(strs) > 1 {
		tmp := make([]string, len(strs))
		for idx, str := range strs {
			tmp[idx], _ = maskNameString(arg, str)
		}
		return strings.Join(tmp, " "), nil
	}

	if l == 2 || l == 3 {
		return overlay(value, strLoop(maskChar, len("**")), 1, 2), nil
	}

	if l > 3 {
		return overlay(value, strLoop(maskChar, len("**")), 1, 3), nil
	}

	return strLoop(maskChar, len("**")), nil
}

func strLoop(str string, length int) string {
	var mask string
	for i := 1; i <= length; i++ {
		mask += str
	}
	return mask
}

func overlay(str string, overlay string, start int, end int) (overlayed string) {
	r := []rune(str)
	l := len([]rune(r))

	if l == 0 {
		return ""
	}

	if start < 0 {
		start = 0
	}
	if start > l {
		start = l
	}
	if end < 0 {
		end = 0
	}
	if end > l {
		end = l
	}
	if start > end {
		tmp := start
		start = end
		end = tmp
	}

	overlayed = ""
	overlayed += string(r[:start])
	overlayed += overlay
	overlayed += string(r[end:])
	return overlayed
}
