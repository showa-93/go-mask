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

	"reflect"
)

func init() {
	defaultMasker = NewMasker()
	defaultMasker.RegisterMaskStringFunc(MaskTypeFilled, defaultMasker.MaskFilledString)
	defaultMasker.RegisterMaskStringFunc(MaskTypeFixed, defaultMasker.MaskFixedString)
	defaultMasker.RegisterMaskStringFunc(MaskTypeHash, defaultMasker.MaskHashString)
	defaultMasker.RegisterMaskIntFunc(MaskTypeRandom, defaultMasker.MaskRandomInt)
	defaultMasker.RegisterMaskFloat64Func(MaskTypeRandom, defaultMasker.MaskRandomFloat64)
	defaultMasker.RegisterMaskAnyFunc(MaskTypeZero, defaultMasker.MaskZero)
}

// Tag name of the field in the structure when masking
const TagName = "mask"

const maskChar = "*"

// Default tag that can be specified as a mask
const (
	MaskTypeFilled = "filled"
	MaskTypeFixed  = "fixed"
	MaskTypeRandom = "random"
	MaskTypeHash   = "hash"
	MaskTypeZero   = "zero"
)

var defaultMasker *Masker

// Function type that must be satisfied to add a custom mask
type (
	MaskStringFunc  func(arg string, value string) (string, error)
	MaskUintFunc    func(arg string, value uint) (uint, error)
	MaskIntFunc     func(arg string, value int) (int, error)
	MaskFloat64Func func(arg string, value float64) (float64, error)
	MaskAnyFunc     func(arg string, value any) (any, error)
)

// Mask returns an object with the mask applied to any given object.
// The function's argument can accept any type, including pointer, map, and slice types, in addition to struct.
// from default masker.
func Mask[T any](target T) (ret T, err error) {
	var v any
	v, err = defaultMasker.Mask(target)
	if err != nil {
		return ret, err
	}

	return v.(T), nil
}

// SetMaskChar changes the character used for masking
// from default masker.
func SetMaskChar(s string) {
	defaultMasker.SetMaskChar(s)
}

// MaskChar returns the current character used for masking.
// from default masker.
func MaskChar() string {
	return defaultMasker.MaskChar()
}

// RegisterMaskField allows you to register a mask tag to be applied to the value of a struct field or map key that matches the fieldName.
// If a mask tag is set on the struct field, it will take precedence.
// from default masker.
func RegisterMaskField(fieldName, maskType string) {
	defaultMasker.RegisterMaskField(fieldName, maskType)
}

// RegisterMaskStringFunc registers a masking function for string values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
// from default masker.
func RegisterMaskStringFunc(maskType string, maskFunc MaskStringFunc) {
	defaultMasker.RegisterMaskStringFunc(maskType, maskFunc)
}

// RegisterMaskIntFunc registers a masking function for int values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
// from default masker.
func RegisterMaskIntFunc(maskType string, maskFunc MaskIntFunc) {
	defaultMasker.RegisterMaskIntFunc(maskType, maskFunc)
}

// RegisterMaskUintFunc registers a masking function for uint values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
// from default masker.
func RegisterMaskUintFunc(maskType string, maskFunc MaskUintFunc) {
	defaultMasker.RegisterMaskUintFunc(maskType, maskFunc)
}

// RegisterMaskFloat64Func registers a masking function for float64 values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
// from default masker.
func RegisterMaskFloat64Func(maskType string, maskFunc MaskFloat64Func) {
	defaultMasker.RegisterMaskFloat64Func(maskType, maskFunc)
}

// RegisterMaskAnyFunc registers a masking function that can be applied to any type.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
// from default masker.
func RegisterMaskAnyFunc(maskType string, maskFunc MaskAnyFunc) {
	defaultMasker.RegisterMaskAnyFunc(maskType, maskFunc)
}

// String masks the given argument string
// from default masker.
func String(tag, value string) (string, error) {
	return defaultMasker.String(tag, value)
}

// Int masks the given argument int
// from default masker.
func Int(tag string, value int) (int, error) {
	return defaultMasker.Int(tag, value)
}

// Uint masks the given argument int
// from default masker.
func Uint(tag string, value uint) (uint, error) {
	return defaultMasker.Uint(tag, value)
}

// Float64 masks the given argument float64
// from default masker.
func Float64(tag string, value float64) (float64, error) {
	return defaultMasker.Float64(tag, value)
}

// structType stores the type information of a structure when caching is enabled
type structType struct {
	value        reflect.Value
	structFields []reflect.StructField
}

// Masker is a struct that defines the masking process.
type Masker struct {
	cache             bool
	mu                sync.RWMutex
	tagName           string
	maskChar          string
	typeToStructCache map[reflect.Type]structType

	maskFieldMap map[string]string

	maskStringFuncKeys  []string
	maskStringFuncMap   map[string]MaskStringFunc
	maskUintFuncKeys    []string
	maskUintFuncMap     map[string]MaskUintFunc
	maskIntFuncKeys     []string
	maskIntFuncMap      map[string]MaskIntFunc
	maskFloat64FuncKeys []string
	maskFloat64FuncMap  map[string]MaskFloat64Func
	maskAnyFuncKeys     []string
	maskAnyFuncMap      map[string]MaskAnyFunc
}

// NewMasker initializes a Masker.
func NewMasker() *Masker {
	m := &Masker{
		tagName:  TagName,
		maskChar: maskChar,

		cache:             true,
		typeToStructCache: make(map[reflect.Type]structType),

		maskFieldMap: make(map[string]string),

		maskStringFuncKeys:  make([]string, 0, 10),
		maskStringFuncMap:   make(map[string]MaskStringFunc),
		maskUintFuncKeys:    make([]string, 0, 10),
		maskUintFuncMap:     make(map[string]MaskUintFunc),
		maskIntFuncKeys:     make([]string, 0, 10),
		maskIntFuncMap:      make(map[string]MaskIntFunc),
		maskFloat64FuncKeys: make([]string, 0, 10),
		maskFloat64FuncMap:  make(map[string]MaskFloat64Func),
		maskAnyFuncKeys:     make([]string, 0, 10),
		maskAnyFuncMap:      make(map[string]MaskAnyFunc),
	}

	return m
}

// SetTagName allows you to change the tag name from "mask" to something else.
func (m *Masker) SetTagName(s string) {
	if s != "" {
		m.tagName = s
	}
}

// SetMaskChar changes the character used for masking
func (m *Masker) SetMaskChar(s string) {
	m.maskChar = s
}

// Cache can be toggled to cache the type information of the struct.
// default true
func (m *Masker) Cache(enable bool) {
	m.cache = enable
}

// MaskChar returns the current character used for masking.
func (m *Masker) MaskChar() string {
	return m.maskChar
}

func (m *Masker) getTag(tag, key string) string {
	if tag != "" {
		return tag
	}
	return m.maskFieldMap[key]
}

// RegisterMaskStringFunc registers a masking function for string values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
func (m *Masker) RegisterMaskStringFunc(maskType string, maskFunc MaskStringFunc) {
	if _, ok := m.maskStringFuncMap[maskType]; !ok {
		m.maskStringFuncKeys = append(m.maskStringFuncKeys, maskType)
	}
	m.maskStringFuncMap[maskType] = maskFunc
}

// RegisterMaskUintFunc registers a masking function for uint values.
// The function will be applied when the uint slice set in the first argument is assigned as a tag to a field in the structure.
func (m *Masker) RegisterMaskUintFunc(maskType string, maskFunc MaskUintFunc) {
	if _, ok := m.maskUintFuncMap[maskType]; !ok {
		m.maskUintFuncKeys = append(m.maskUintFuncKeys, maskType)
	}
	m.maskUintFuncMap[maskType] = maskFunc
}

// RegisterMaskIntFunc registers a masking function for int values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
func (m *Masker) RegisterMaskIntFunc(maskType string, maskFunc MaskIntFunc) {
	if _, ok := m.maskIntFuncMap[maskType]; !ok {
		m.maskIntFuncKeys = append(m.maskIntFuncKeys, maskType)
	}
	m.maskIntFuncMap[maskType] = maskFunc
}

// RegisterMaskFloat64Func registers a masking function for float64 values.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
func (m *Masker) RegisterMaskFloat64Func(maskType string, maskFunc MaskFloat64Func) {
	if _, ok := m.maskFloat64FuncMap[maskType]; !ok {
		m.maskFloat64FuncKeys = append(m.maskFloat64FuncKeys, maskType)
	}
	m.maskFloat64FuncMap[maskType] = maskFunc
}

// RegisterMaskAnyFunc registers a masking function that can be applied to any type.
// The function will be applied when the string set in the first argument is assigned as a tag to a field in the structure.
func (m *Masker) RegisterMaskAnyFunc(maskType string, maskFunc MaskAnyFunc) {
	if _, ok := m.maskAnyFuncMap[maskType]; !ok {
		m.maskAnyFuncKeys = append(m.maskAnyFuncKeys, maskType)
	}
	m.maskAnyFuncMap[maskType] = maskFunc
}

// RegisterMaskField allows you to register a mask tag to be applied to the value of a struct field or map key that matches the fieldName.
// If a mask tag is set on the struct field, it will take precedence.
func (m *Masker) RegisterMaskField(fieldName, maskType string) {
	m.maskFieldMap[fieldName] = maskType
}

// String masks the given argument string
func (m *Masker) String(tag, value string) (string, error) {
	if tag != "" {
		for _, mt := range m.maskStringFuncKeys {
			if strings.HasPrefix(tag, mt) {
				return m.maskStringFuncMap[mt](tag[len(mt):], value)
			}
		}
		if ok, v, err := m.maskAny(tag, value); ok {
			return v.(string), err
		}
	}

	return value, nil
}

// Uint masks the given argument uint
func (m *Masker) Uint(tag string, value uint) (uint, error) {
	if tag != "" {
		for _, mt := range m.maskUintFuncKeys {
			if strings.HasPrefix(tag, mt) {
				return m.maskUintFuncMap[mt](tag[len(mt):], value)
			}
		}
		if ok, v, err := m.maskAny(tag, value); ok {
			return v.(uint), err
		}
	}

	return value, nil
}

// Int masks the given argument int
func (m *Masker) Int(tag string, value int) (int, error) {
	if tag != "" {
		for _, mt := range m.maskIntFuncKeys {
			if strings.HasPrefix(tag, mt) {
				return m.maskIntFuncMap[mt](tag[len(mt):], value)
			}
		}
		if ok, v, err := m.maskAny(tag, value); ok {
			return v.(int), err
		}
	}

	return value, nil
}

// Float64 masks the given argument float64
func (m *Masker) Float64(tag string, value float64) (float64, error) {
	if tag != "" {
		for _, mt := range m.maskFloat64FuncKeys {
			if strings.HasPrefix(tag, mt) {
				return m.maskFloat64FuncMap[mt](tag[len(mt):], value)
			}
		}
		if ok, v, err := m.maskAny(tag, value); ok {
			return v.(float64), err
		}
	}

	return value, nil
}

func (m *Masker) maskAny(tag string, value any) (bool, any, error) {
	if tag != "" {
		for _, mt := range m.maskAnyFuncKeys {
			if strings.HasPrefix(tag, mt) {
				v, err := m.maskAnyFuncMap[mt](tag[len(mt):], value)
				return true, v, err
			}
		}
	}

	return false, value, nil
}

func (m *Masker) maskAnyValue(tag string, value reflect.Value) (bool, reflect.Value, error) {
	if tag != "" {
		for _, mt := range m.maskAnyFuncKeys {
			if strings.HasPrefix(tag, mt) {
				v, err := m.maskAnyFuncMap[mt](tag[len(mt):], value.Interface())
				return true, reflect.ValueOf(v), err
			}
		}
	}

	return false, value, nil
}

// MaskFilledString masks the string length of the value with the same length.
// If you pass a number like "2" to arg, it masks with the length of the number.(**)
func (m *Masker) MaskFilledString(arg, value string) (string, error) {
	if arg != "" {
		count, err := strconv.Atoi(arg)
		if err != nil {
			return "", err
		}

		return strings.Repeat(m.MaskChar(), count), nil
	}

	return strings.Repeat(m.MaskChar(), utf8.RuneCountInString(value)), nil
}

// MaskFixedString masks with a fixed length (8 characters).
func (m *Masker) MaskFixedString(arg, value string) (string, error) {
	return strings.Repeat(m.MaskChar(), 8), nil
}

// MaskHashString masks and hashes (sha1) a string.
func (m *Masker) MaskHashString(arg, value string) (string, error) {
	hash := sha1.Sum(([]byte)(value))
	return hex.EncodeToString(hash[:]), nil
}

// MaskRandomInt converts an integer (int) into a random number.
// For example, if you pass "100" as the arg, it sets a random number in the range of 0-99.
func (m *Masker) MaskRandomInt(arg string, value int) (int, error) {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return 0, err
	}

	return rand.Intn(n), nil
}

// MaskRandomFloat64 converts a float64 to a random number.
// For example, if you pass "100.3" to arg, it sets a random number in the range of 0.000 to 99.999.
func (m *Masker) MaskRandomFloat64(arg string, value float64) (float64, error) {
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

// MaskZero converts the value to its type's zero value.
func (m *Masker) MaskZero(arg string, value any) (any, error) {
	return reflect.Zero(reflect.TypeOf(value)).Interface(), nil
}

// Mask returns an object with the mask applied to any given object.
// The function's argument can accept any type, including pointer, map, and slice types, in addition to struct.
func (m *Masker) Mask(target any) (ret any, err error) {
	rv, err := m.mask(reflect.ValueOf(target), "", reflect.Value{})
	if err != nil {
		return ret, err
	}

	return rv.Interface(), nil
}

func (m *Masker) mask(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if ok, v, err := m.maskAnyValue(tag, rv); ok {
		return v, err
	}
	switch rv.Type().Kind() {
	case reflect.Interface:
		return m.maskInterface(rv, tag, mp)
	case reflect.Ptr:
		return m.maskPtr(rv, tag, mp)
	case reflect.Struct:
		return m.maskStruct(rv, tag, mp)
	case reflect.Array:
		return m.maskSlice(rv, tag, mp)
	case reflect.Slice:
		if rv.IsNil() {
			return reflect.Zero(rv.Type()), nil
		}
		return m.maskSlice(rv, tag, mp)
	case reflect.Map:
		return m.maskMap(rv, tag, mp)
	case reflect.String:
		return m.maskString(rv, tag, mp)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return m.maskInt(rv, tag, mp)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return m.maskUint(rv, tag, mp)
	case reflect.Float32, reflect.Float64:
		return m.maskfloat(rv, tag, mp)
	default:
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}
}

func (m *Masker) maskInterface(rv reflect.Value, tag string, _ reflect.Value) (reflect.Value, error) {
	if rv.IsNil() {
		return reflect.Zero(rv.Type()), nil
	}

	mp := reflect.New(rv.Type()).Elem()
	rv2, err := m.mask(reflect.ValueOf(rv.Interface()), tag, reflect.Value{})
	if err != nil {
		return reflect.Value{}, err
	}
	mp.Set(rv2)

	return mp, nil
}

func (m *Masker) maskPtr(rv reflect.Value, tag string, _ reflect.Value) (reflect.Value, error) {
	if rv.IsNil() {
		return reflect.Zero(rv.Type()), nil
	}

	mp := reflect.New(rv.Type().Elem())
	rv2, err := m.mask(rv.Elem(), tag, mp.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	mp.Elem().Set(rv2)

	return mp, nil
}

func (m *Masker) maskStruct(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsZero() {
		return reflect.Zero(rv.Type()), nil
	}

	rt := rv.Type()
	var st structType
	if m.cache {
		m.mu.RLock()
		var ok bool
		st, ok = m.typeToStructCache[rt]
		m.mu.RUnlock()
		if !ok {
			m.mu.Lock()
			st.value = reflect.New(rt).Elem()
			for i := 0; i < rt.NumField(); i++ {
				st.structFields = append(st.structFields, rt.Field(i))
			}
			m.typeToStructCache[rt] = st
			m.mu.Unlock()
		}
		if !mp.IsValid() {
			mp = st.value
		}
	} else {
		if !mp.IsValid() {
			mp = reflect.New(rt).Elem()
		}
	}

	for i := 0; i < rt.NumField(); i++ {
		var field reflect.StructField
		if m.cache {
			field = st.structFields[i]
		} else {
			field = rt.Field(i)
		}
		// skip private field
		if field.PkgPath != "" {
			continue
		}
		tag := field.Tag.Get(m.tagName)
		switch field.Type.Kind() {
		case reflect.String:
			s, err := m.String(m.getTag(tag, field.Name), rv.Field(i).String())
			if err != nil {
				return reflect.Value{}, err
			}
			mp.Field(i).SetString(s)
		default:
			rvf, err := m.mask(rv.Field(i), m.getTag(tag, field.Name), mp.Field(i))
			if err != nil {
				return reflect.Value{}, err
			}
			mp.Field(i).Set(rvf)
		}
	}

	return mp, nil
}

func (m *Masker) maskSlice(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	var rv2 reflect.Value

	if rt := rv.Type(); rt.Kind() == reflect.Array {
		rv2 = reflect.New(rt).Elem()
	} else {
		rv2 = reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
	}
	for i := 0; i < rv.Len(); i++ {
		value := rv.Index(i)
		switch rv.Type().Elem().Kind() {
		case reflect.String:
			rvf, err := m.String(tag, value.String())
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetString(rvf)
		case reflect.Int:
			rvf, err := m.Int(tag, int(value.Int()))
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetInt(int64(rvf))
		case reflect.Float64:
			rvf, err := m.Float64(tag, value.Float())
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetFloat(rvf)
		case reflect.Uint:
			rvf, err := m.Uint(tag, uint(value.Uint()))
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).SetUint(uint64(rvf))
		default:
			rvf, err := m.mask(value, tag, rv2.Index(i))
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.Index(i).Set(rvf)
		}
	}

	if mp.IsValid() {
		mp.Set(rv2)
		return mp, nil
	}

	return rv2, nil
}

func (m *Masker) maskMap(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if rv.IsNil() {
		return reflect.Zero(rv.Type()), nil
	}

	switch rv.Type().Key().Kind() {
	case reflect.String:
		rv2, err := m.maskStringKeyMap(rv, tag)
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

	rv2, err := m.maskAnyKeyMap(rv, tag)
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.Set(rv2)
		return mp, nil
	}

	return rv2, nil
}

func (m *Masker) maskAnyKeyMap(rv reflect.Value, tag string) (reflect.Value, error) {
	rv2 := reflect.MakeMapWithSize(rv.Type(), rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		key, value := iter.Key(), iter.Value()
		rf, err := m.mask(value, tag, reflect.Value{})
		if err != nil {
			return reflect.Value{}, err
		}
		rv2.SetMapIndex(key, rf)
	}

	return rv2, nil
}

func (m *Masker) maskStringKeyMap(rv reflect.Value, tag string) (reflect.Value, error) {
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		mm := make(map[string]string, rv.Len())
		for k, v := range rv.Convert(reflect.TypeOf(mm)).Interface().(map[string]string) {
			rvf, err := m.String(m.getTag(tag, k), v)
			if err != nil {
				return reflect.Value{}, err
			}
			mm[k] = rvf
		}

		return reflect.ValueOf(mm).Convert(rv.Type()), nil
	case reflect.Int:
		mm := make(map[string]int, rv.Len())
		for k, v := range rv.Convert(reflect.TypeOf(mm)).Interface().(map[string]int) {
			rvf, err := m.Int(m.getTag(tag, k), v)
			if err != nil {
				return reflect.Value{}, err
			}
			mm[k] = rvf
		}
		return reflect.ValueOf(mm).Convert(rv.Type()), nil
	case reflect.Float64:
		mm := make(map[string]float64, rv.Len())
		for k, v := range rv.Convert(reflect.TypeOf(mm)).Interface().(map[string]float64) {
			rvf, err := m.Float64(m.getTag(tag, k), v)
			if err != nil {
				return reflect.Value{}, err
			}
			mm[k] = rvf
		}

		return reflect.ValueOf(mm).Convert(rv.Type()), nil
	default:
		rv2 := reflect.MakeMapWithSize(rv.Type(), rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key, value := iter.Key(), iter.Value()
			rf, err := m.mask(value, m.getTag(tag, key.String()), reflect.Value{})
			if err != nil {
				return reflect.Value{}, err
			}
			rv2.SetMapIndex(key, rf)
		}
		return rv2, nil
	}

}

func (m *Masker) maskString(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	sp, err := m.String(tag, rv.String())
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.SetString(sp)
		return mp, nil
	}

	return valueOfString(sp), nil
}

func valueOfString(s string) reflect.Value {
	return reflect.ValueOf(&s).Elem()
}

func (m *Masker) maskInt(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	ip, err := m.Int(tag, int(rv.Int()))
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

func (m *Masker) maskUint(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	ip, err := m.Uint(tag, uint(rv.Uint()))
	if err != nil {
		return reflect.Value{}, err
	}
	if mp.IsValid() {
		mp.SetUint(uint64(ip))
		return mp, nil
	}

	if rv.Type().Kind() != reflect.Uint {
		return reflect.ValueOf(&ip).Elem().Convert(rv.Type()), nil
	}

	return reflect.ValueOf(&ip).Elem(), nil
}

func (m *Masker) maskfloat(rv reflect.Value, tag string, mp reflect.Value) (reflect.Value, error) {
	if tag == "" {
		if mp.IsValid() {
			mp.Set(rv)
			return mp, nil
		}
		return rv, nil
	}

	fp, err := m.Float64(tag, rv.Float())
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
