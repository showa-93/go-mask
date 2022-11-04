// 日付マスキング：有効なランダムな日付+-
// ハッシュ化
// 正規表現
// zero値にする
package maskgo

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/goccy/go-reflect"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const tagName = "mask"

const (
	MaskTypeFilled = "filled"
	MaskTypeRandom = "random"
)

type storeStruct struct {
	rv           reflect.Value
	structFields []reflect.StructField
}

type (
	maskStringFunc func(arg, value string) (string, error)
	maskIntFunc    func(arg string, value int) (int, error)
)

var (
	typeToStruct      sync.Map
	maskChar                                    = "*"
	maskStringFuncMap map[string]maskStringFunc = map[string]maskStringFunc{
		MaskTypeFilled: maskFilledString,
	}
	maskIntFuncMap map[string]maskIntFunc = map[string]maskIntFunc{
		MaskTypeRandom: maskRandomInt,
	}
)

func RegisterMaskFunc(maskType string, f maskStringFunc) {
	maskStringFuncMap[maskType] = f
}

func MaskString(tag, value string) (string, error) {
	if tag != "" {
		for mt, maskStringFunc := range maskStringFuncMap {
			if strings.HasPrefix(tag, mt) {
				return maskStringFunc(tag[len(mt):], value)
			}
		}
	}

	return value, nil
}

func maskFilledString(arg, value string) (string, error) {
	return strings.Repeat(maskChar, utf8.RuneCountInString(value)), nil
}

func MaskInt(tag string, value int) (int, error) {
	if tag != "" {
		for mt, maskIntFunc := range maskIntFuncMap {
			if strings.HasPrefix(tag, mt) {
				return maskIntFunc(tag[len(mt):], value)
			}
		}
	}

	return value, nil
}

func maskRandomInt(arg string, value int) (int, error) {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return 0, err
	}

	return rand.Intn(n), nil
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
	switch rv.Type().Kind() {
	case reflect.Ptr:
		return maskPtr(rv, tag)
	case reflect.Struct:
		return maskStruct(rv, tag)
	case reflect.Slice:
		return maskSlice(rv, tag)
	case reflect.String:
		return maskString(rv, tag)
	case reflect.Int:
		return maskInt(rv, tag)
	default:
		return rv, nil
	}
}

func maskPtr(rv reflect.Value, tag string) (reflect.Value, error) {
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
}

func maskStruct(rv reflect.Value, tag string) (reflect.Value, error) {
	if rv.IsZero() {
		return rv, nil
	}

	var ss storeStruct
	rt := rv.Type()
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
}

func maskSlice(rv reflect.Value, tag string) (reflect.Value, error) {
	if rv.IsZero() {
		return rv, nil
	}

	rv2 := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		for i, str := range rv.Interface().([]string) {
			rvf, err := MaskString(tag, str)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetString(rvf)
		}
	case reflect.Int:
		for i, v := range rv.Interface().([]int) {
			rvf, err := MaskInt(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetInt(int64(rvf))
		}
	default:
		for i := 0; i < rv.Len(); i++ {
			rf, err := mask(rv.Index(i), tag)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).Set(rf)
		}
	}

	return rv2, nil
}

func maskString(rv reflect.Value, tag string) (reflect.Value, error) {
	sp, err := MaskString(tag, rv.String())
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(&sp).Elem(), nil
}

func maskInt(rv reflect.Value, tag string) (reflect.Value, error) {
	ip, err := MaskInt(tag, rv.Interface().(int))
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(&ip).Elem(), nil
}
