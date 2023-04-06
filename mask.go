package mask

import (
	"crypto/sha1"
	"encoding/hex"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/goccy/go-reflect"
)

const tagName = "mask"

const (
	MaskTypeFilled = "filled"
	MaskTypeRandom = "random"
	MaskTypeHash   = "hash"
	MaskTypeZero   = "zero"
)

type storeStruct struct {
	mv           reflect.Value
	structFields []reflect.StructField
}

type (
	MaskStringFunc  func(arg string, value string) (string, error)
	MaskIntFunc     func(arg string, value int) (int, error)
	MaskFloat64Func func(arg string, value float64) (float64, error)
	MaskAnyFunc     func(arg string, value any) (any, error)
)

var (
	maskChar          = "*"
	typeToStructMap   sync.Map
	maskStringFuncMap = map[string]MaskStringFunc{
		MaskTypeFilled: maskFilledString,
		MaskTypeHash:   maskHashString,
	}
	maskIntFuncMap = map[string]MaskIntFunc{
		MaskTypeRandom: maskRandomInt,
	}
	maskFloat64FuncMap = map[string]MaskFloat64Func{
		MaskTypeRandom: maskRandomFloat64,
	}
	maskAnyFuncMap = map[string]MaskAnyFunc{
		MaskTypeZero: maskZero,
	}
)

func SetMaskChar(s string) {
	maskChar = s
}

func MaskChar() string {
	return maskChar
}

func RegisterMaskStringFunc(maskType string, maskFunc MaskStringFunc) {
	maskStringFuncMap[maskType] = maskFunc
}

func RegisterMaskIntFunc(maskType string, maskFunc MaskIntFunc) {
	maskIntFuncMap[maskType] = maskFunc
}

func RegisterMaskFloat64Func(maskType string, maskFunc MaskFloat64Func) {
	maskFloat64FuncMap[maskType] = maskFunc
}

func RegisterMaskAnyFunc(maskType string, maskFunc MaskAnyFunc) {
	maskAnyFuncMap[maskType] = maskFunc
}

func String(tag, value string) (string, error) {
	if tag != "" {
		if ok, v, err := maskAny(tag, value); ok {
			return v.(string), err
		}
		for mt, maskStringFunc := range maskStringFuncMap {
			if strings.HasPrefix(tag, mt) {
				return maskStringFunc(tag[len(mt):], value)
			}
		}
	}

	return value, nil
}

func Int(tag string, value int) (int, error) {
	if tag != "" {
		if ok, v, err := maskAny(tag, value); ok {
			return v.(int), err
		}
		for mt, maskIntFunc := range maskIntFuncMap {
			if strings.HasPrefix(tag, mt) {
				return maskIntFunc(tag[len(mt):], value)
			}
		}
	}

	return value, nil
}

func Float64(tag string, value float64) (float64, error) {
	if tag != "" {
		if ok, v, err := maskAny(tag, value); ok {
			return v.(float64), err
		}
		for mt, maskFloat64Func := range maskFloat64FuncMap {
			if strings.HasPrefix(tag, mt) {
				return maskFloat64Func(tag[len(mt):], value)
			}
		}
	}

	return value, nil
}

func maskAny(tag string, value any) (bool, any, error) {
	if tag != "" {
		for mt, maskAnyFunc := range maskAnyFuncMap {
			if strings.HasPrefix(tag, mt) {
				v, err := maskAnyFunc(tag[len(mt):], value)
				return true, v, err
			}
		}
	}

	return false, value, nil
}

func maskAnyValue(tag string, value reflect.Value) (bool, reflect.Value, error) {
	if tag != "" {
		for mt, maskAnyFunc := range maskAnyFuncMap {
			if strings.HasPrefix(tag, mt) {
				v, err := maskAnyFunc(tag[len(mt):], value.Interface())
				return true, reflect.ValueOf(v), err
			}
		}
	}

	return false, value, nil
}

func maskFilledString(arg, value string) (string, error) {
	if arg != "" {
		count, err := strconv.Atoi(arg)
		if err != nil {
			return "", err
		}

		return strings.Repeat(MaskChar(), count), nil
	}

	return strings.Repeat(MaskChar(), utf8.RuneCountInString(value)), nil
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
	x := float64(int(rand.Float64() * float64(i) * dd))

	return x / dd, nil
}

func maskZero(arg string, value any) (any, error) {
	return reflect.Zero(reflect.TypeOf(value)).Interface(), nil
}

func Mask(target any) (any, error) {
	if target == nil {
		return target, nil
	}
	rv, err := mask(reflect.ValueOf(target), "", reflect.Value{})
	if err != nil {
		return nil, err
	}

	return rv.Interface(), nil
}

func mask(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if ok, v, err := maskAnyValue(tag, rv); ok {
		return v, err
	}
	switch rv.Type().Kind() {
	case reflect.Interface:
		return maskInterface(rv, tag, mp)
	case reflect.Ptr:
		return maskPtr(rv, tag, mp)
	case reflect.Struct:
		return maskStruct(rv, tag, mp)
	case reflect.Slice:
		return maskSlice(rv, tag, mp)
	case reflect.Map:
		return maskMap(rv, tag, mp)
	case reflect.String:
		return maskString(rv, tag, mp)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return maskInt(rv, tag, mp)
	case reflect.Float32, reflect.Float64:
		return maskfloat(rv, tag, mp)
	default:
		if mp.IsValid() {
			mp.Set(rv)
		}
		return rv, nil
	}
}

func maskInterface(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsNil() {
		return reflect.Zero(rv.Type()), nil
	}

	if !mp.IsValid() {
		mp = reflect.New(rv.Type()).Elem()
	}
	rv2, err := mask(reflect.ValueOf(rv.Interface()), tag, reflect.Value{})
	if err != nil {
		return reflect.Value{}, err
	}
	mp.Set(rv2)

	return mp, nil
}

func maskPtr(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsNil() {
		return reflect.Zero(rv.Type()), nil
	}

	if !mp.IsValid() {
		mp = reflect.New(rv.Type().Elem())
	}
	_, err := mask(rv.Elem(), tag, mp.Elem())
	if err != nil {
		return reflect.Value{}, err
	}

	return mp, nil
}

func maskStruct(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsZero() {
		return reflect.Zero(rv.Type()), nil
	}

	var ss storeStruct
	rt := rv.Type()
	if storeValue, ok := typeToStructMap.Load(rt.String()); ok {
		ss = storeValue.(storeStruct)
		if mp.IsValid() {
			ss.mv = mp
		}
	} else {
		if mp.IsValid() {
			ss.mv = mp
		} else {
			ss.mv = reflect.New(rt).Elem()
		}
		ss.structFields = make([]reflect.StructField, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			ss.structFields[i] = rt.Field(i)
		}

		typeToStructMap.Store(rt.String(), ss)
	}

	for i := 0; i < rv.NumField(); i++ {
		if ss.structFields[i].PkgPath != "" {
			continue
		}
		vTag := ss.structFields[i].Tag.Get(tagName)
		rvf, err := mask(rv.Field(i), vTag, reflect.Value{})
		if err != nil {
			return reflect.Value{}, err
		}
		if rvf.IsValid() {
			ss.mv.Field(i).Set(rvf)
		}
	}

	return ss.mv, nil
}

func maskSlice(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsZero() {
		return reflect.Zero(rv.Type()), nil
	}

	rv2 := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		for i, str := range rv.Interface().([]string) {
			rvf, err := String(tag, str)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetString(rvf)
		}
	case reflect.Int:
		for i, v := range rv.Interface().([]int) {
			rvf, err := Int(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetInt(int64(rvf))
		}
	case reflect.Float64:
		for i, v := range rv.Interface().([]float64) {
			rvf, err := Float64(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetFloat(rvf)
		}
	default:
		for i := 0; i < rv.Len(); i++ {
			_, err := mask(rv.Index(i), tag, rv2.Index(i))
			if err != nil {
				return reflect.Value{}, err
			}
		}
	}

	if mp.IsValid() {
		mp.Set(rv2)
	}

	return rv2, nil
}

func maskMap(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsNil() {
		return reflect.Zero(rv.Type()), nil
	}

	switch rv.Type().Key().Kind() {
	case reflect.String:
		rv2, err := maskStringKeyMap(rv, tag)
		if err != nil {
			return reflect.Value{}, err
		}
		if rv2.IsValid() {
			if mp.IsValid() {
				mp.Set(rv2)
			}
			return rv2, nil
		}
	}

	rv2, err := maskAnyKeykMap(rv, tag)
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.Set(rv2)
	}

	return rv2, nil
}

func maskAnyKeykMap(rv reflect.Value, tag string) (reflect.Value, error) {
	rv2 := reflect.MakeMapWithSize(rv.Type(), rv.Len())
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		iter := rv.MapRange()
		for iter.Next() {
			key, value := iter.Key(), iter.Value()
			rvf, err := String(tag, value.String())
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(reflect.ToValue(key), reflect.ValueOf(rvf))
		}
	case reflect.Int:
		iter := rv.MapRange()
		for iter.Next() {
			key, value := iter.Key(), iter.Value()
			rvf, err := Int(tag, int(value.Int()))
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(reflect.ToValue(key), reflect.ValueOf(rvf))
		}
	case reflect.Float64:
		iter := rv.MapRange()
		for iter.Next() {
			key, value := iter.Key(), iter.Value()
			rvf, err := Float64(tag, value.Float())
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(reflect.ToValue(key), reflect.ValueOf(rvf))
		}
	default:
		iter := rv.MapRange()
		for iter.Next() {
			key, value := iter.Key(), iter.Value()
			rf, err := mask(reflect.ToValue(value), tag, reflect.Value{})
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(reflect.ToValue(key), rf)
		}
	}

	return rv2, nil
}

func maskStringKeyMap(rv reflect.Value, tag string) (reflect.Value, error) {
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		m := make(map[string]string, rv.Len())
		for k, v := range rv.Interface().(map[string]string) {
			rvf, err := String(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			m[k] = rvf
		}

		return reflect.ValueOf(m), nil
	case reflect.Int:
		m := make(map[string]int, rv.Len())
		for k, v := range rv.Interface().(map[string]int) {
			rvf, err := Int(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			m[k] = rvf
		}
		return reflect.ValueOf(m), nil
	case reflect.Float64:
		m := make(map[string]float64, rv.Len())
		for k, v := range rv.Interface().(map[string]float64) {
			rvf, err := Float64(tag, v)
			if err != nil {
				return reflect.Value{}, err
			}
			m[k] = rvf
		}

		return reflect.ValueOf(m), nil
	}

	return reflect.Value{}, nil
}

func maskString(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	sp, err := String(tag, rv.String())
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.SetString(sp)
		return mp, nil
	}

	return reflect.ValueOf(&sp).Elem(), nil
}

func maskInt(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	ip, err := Int(tag, int(rv.Int()))
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.SetInt(int64(ip))
		return mp, nil
	}

	if rv.Type().Kind() != reflect.Int {
		return reflect.ValueOf(&ip).Elem().Convert(rv.Type()), nil
	}

	return reflect.ValueOf(&ip).Elem(), nil
}

func maskfloat(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	fp, err := Float64(tag, rv.Float())
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.SetFloat(fp)
		return mp, nil
	}

	if rv.Type().Kind() != reflect.Float64 {
		return reflect.ValueOf(&fp).Elem().Convert(rv.Type()), nil
	}

	return reflect.ValueOf(&fp).Elem(), nil
}
