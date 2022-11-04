// 日付マスキング：有効なランダムな日付+-
// 正規表現
// zero値にする
package maskgo

import (
	"crypto/sha1"
	"encoding/hex"
	"math"
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
	MaskTypeHash   = "hash"
)

type storeStruct struct {
	rv           reflect.Value
	structFields []reflect.StructField
}

type (
	maskStringFunc  func(arg, value string) (string, error)
	maskIntFunc     func(arg string, value int) (int, error)
	maskFloat64Func func(arg string, value float64) (float64, error)
)

var (
	typeToStruct      sync.Map
	maskChar                                    = "*"
	maskStringFuncMap map[string]maskStringFunc = map[string]maskStringFunc{
		MaskTypeFilled: maskFilledString,
		MaskTypeHash:   maskHashString,
	}
	maskIntFuncMap map[string]maskIntFunc = map[string]maskIntFunc{
		MaskTypeRandom: maskRandomInt,
	}
	maskFloat64FuncMap map[string]maskFloat64Func = map[string]maskFloat64Func{
		MaskTypeRandom: maskRandomFloat64,
	}
)

func RegisterMaskStringFunc(maskType string, f maskStringFunc) {
	maskStringFuncMap[maskType] = f
}

func RegisterMaskIntFunc(maskType string, f maskIntFunc) {
	maskIntFuncMap[maskType] = f
}

func RegisterMaskFloat64Func(maskType string, f maskFloat64Func) {
	maskFloat64FuncMap[maskType] = f
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

func MaskFloat64(tag string, value float64) (float64, error) {
	if tag != "" {
		for mt, maskFloat64Func := range maskFloat64FuncMap {
			if strings.HasPrefix(tag, mt) {
				return maskFloat64Func(tag[len(mt):], value)
			}
		}
	}

	return value, nil
}

func maskFilledString(arg, value string) (string, error) {
	return strings.Repeat(maskChar, utf8.RuneCountInString(value)), nil
}

func maskHashString(arg, value string) (string, error) {
	hash := sha1.Sum(([]byte)(value))
	return hex.EncodeToString(hash[:]), nil
}

func maskRandomInt(arg string, value int) (int, error) {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return 0, err
	}

	return rand.Intn(n), nil
}

func maskRandomFloat64(arg string, value float64) (float64, error) {
	var (
		i, d int
		err  error
	)
	digits := strings.Split(arg, ".")
	if len(digits) > 0 {
		if i, err = strconv.Atoi(digits[0]); err != nil {
			return 0, err
		}
	}
	if len(digits) == 2 {
		if d, err = strconv.Atoi(digits[1]); err != nil {
			return 0, err
		}
	}

	dd := math.Pow10(d)
	x := float64(int(rand.Float64() * math.Pow10(i) * dd))

	return x / dd, nil
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
	case reflect.Map:
		return maskMap(rv, tag)
	case reflect.String:
		return maskString(rv, tag)
	case reflect.Int:
		return maskInt(rv, tag)
	case reflect.Float64:
		return maskfloat64(rv, tag)
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
	case reflect.Float64:
		for i, v := range rv.Interface().([]float64) {
			rvf, err := MaskFloat64(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetFloat(rvf)
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

func maskMap(rv reflect.Value, tag string) (reflect.Value, error) {
	if rv.IsNil() {
		return rv, nil
	}

	switch rv.Type().Key().Kind() {
	case reflect.String:
		rv2, err := maskStringKeyMap(rv, tag)
		if err != nil {
			return reflect.Value{}, err
		}
		if rv2.IsValid() {
			return rv2, nil
		}
	}

	return maskAnyKeykMap(rv, tag)
}

func maskAnyKeykMap(rv reflect.Value, tag string) (reflect.Value, error) {
	rv2 := reflect.MakeMapWithSize(rv.Type(), rv.Len())
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		for _, key := range rv.MapKeys() {
			rvf, err := MaskString(tag, rv.MapIndex(key).String())
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(key, reflect.ValueOf(rvf))
		}
	case reflect.Int:
		for _, key := range rv.MapKeys() {
			rvf, err := MaskInt(tag, int(rv.MapIndex(key).Int()))
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(key, reflect.ValueOf(rvf))
		}
	case reflect.Float64:
		for _, key := range rv.MapKeys() {
			rvf, err := MaskFloat64(tag, rv.MapIndex(key).Float())
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(key, reflect.ValueOf(rvf))
		}
	}

	return rv2, nil
}

func maskStringKeyMap(rv reflect.Value, tag string) (reflect.Value, error) {
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		m := make(map[string]string, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value().String()
			rvf, err := MaskString(tag, value)
			if err != nil {
				return reflect.Value{}, err
			}
			m[key.String()] = rvf
		}
		return reflect.ValueOf(m), nil
	case reflect.Int:
		m := make(map[string]int, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := int(iter.Value().Int())
			rvf, err := MaskInt(tag, value)
			if err != nil {
				return reflect.Value{}, err
			}
			m[key.String()] = rvf
		}
		return reflect.ValueOf(m), nil
	case reflect.Float64:
		m := make(map[string]float64, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value().Float()
			rvf, err := MaskFloat64(tag, value)
			if err != nil {
				return reflect.Value{}, err
			}
			m[key.String()] = rvf
		}
		return reflect.ValueOf(m), nil
	}

	return reflect.Value{}, nil
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

func maskfloat64(rv reflect.Value, tag string) (reflect.Value, error) {
	fp, err := MaskFloat64(tag, rv.Interface().(float64))
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(&fp).Elem(), nil
}
