package hi

import (
	"errors"
	"strconv"

	"github.com/nbcx/go-kit/to"
)

// todo: 待整理

func (c *Context) Form(key string, defaultValue ...any) (value to.Value) {
	c.initFormCache()
	values, ok := c.formCache[key]
	if !ok || len(values) == 0 {
		num := len(defaultValue)
		if num == 0 {
			return
		}
		return to.ValueF(defaultValue[0])
	}
	return to.ValueF(values[0])
}

func (c *Context) FormArray(key string, defaultValue ...any) (nValues to.Values[string]) {
	c.initFormCache()
	values, ok := c.formCache[key]
	if ok {
		return to.ValuesF(values)
	}
	if len(defaultValue) == 0 {
		return
	}
	return to.ValuesF(to.ValuesF(defaultValue).String())
}

func (c *Context) FormMap(key string, defaultValue ...any) (nValues Values) {
	return nil
}

func (c *Context) Query2(key string, defaultValue ...string) (value to.Value) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	if !ok || len(values) == 0 {
		num := len(defaultValue)
		if num == 0 {
			return
		}
		return to.ValueF(defaultValue[0])
	}
	return to.ValueF(values[0])
}

func (c *Context) Query2Array(key string, defaultValue ...any) (nValues to.Values[string]) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	if ok {
		return to.ValuesF(values)
	}
	if len(defaultValue) == 0 {
		return
	}
	return to.ValuesF(to.ValuesF(defaultValue).String())
}

type Values []Value

func (v Values) Array() (i []string) {
	for _, vv := range v {
		i = append(i, string(vv))
	}
	return
}

func (v Values) ArrayInt() (i []int) {
	for _, v := range v {
		i = append(i, v.Int())
	}
	return
}

func (v Values) ArrayInt32() (i []int32) {
	for _, v := range v {
		i = append(i, v.Int32())
	}
	return
}

func (v Values) first() (val Value, ok bool) {
	if len(v) == 0 {
		return val, false
	}
	return v[0], true
}

func (v Values) String() string {
	if f, ok := v.first(); ok {
		return f.String()
	}
	return ""
}

// GetInt returns input as an int or the default value while it's present and input is blank
func (v Values) Int() (i int) {
	i, _ = v.IntE()
	return
}

func (v Values) IntE() (int, error) {
	if f, ok := v.first(); ok {
		return f.IntE()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Int8() int8 {
	i, _ := v.Int8E()
	return i
}

// Int8E return input as an int8 or the default value while it's present and input is blank
func (v Values) Int8E() (int8, error) {
	if f, ok := v.first(); ok {
		return f.Int8E()
	}
	return 0, errors.New("does not exist")
}

// GetUint8 return input as an uint8 or the default value while it's present and input is blank
func (v Values) Uint8() (i uint8) {
	i, _ = v.Uint8E()
	return
}

// GetUint8 return input as an uint8 or the default value while it's present and input is blank
func (v Values) Uint8E() (uint8, error) {
	if f, ok := v.first(); ok {
		return f.Uint8E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Int16() (i int16) {
	i, _ = v.Int16E()
	return
}

// GetInt16 returns input as an int16 or the default value while it's present and input is blank
func (v Values) Int16E() (int16, error) {
	if f, ok := v.first(); ok {
		return f.Int16E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Uint16() (i uint16) {
	i, _ = v.Uint16E()
	return
}

// GetUint16 returns input as an uint16 or the default value while it's present and input is blank
func (v Values) Uint16E() (uint16, error) {
	if f, ok := v.first(); ok {
		return f.Uint16E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Int32() (i int32) {
	i, _ = v.Int32E()
	return
}

// GetInt32 returns input as an int32 or the default value while it's present and input is blank
func (v Values) Int32E() (int32, error) {
	if f, ok := v.first(); ok {
		return f.Int32E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Uint32() (i uint32) {
	i, _ = v.Uint32E()
	return
}

// GetUint32 returns input as an uint32 or the default value while it's present and input is blank
func (v Values) Uint32E() (uint32, error) {
	if f, ok := v.first(); ok {
		return f.Uint32E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Int64() (i int64) {
	i, _ = v.Int64E()
	return
}

// GetInt64 returns input value as int64 or the default value while it's present and input is blank.
func (v Values) Int64E() (int64, error) {
	if f, ok := v.first(); ok {
		return f.Int64E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Uint64() (i uint64) {
	i, _ = v.Uint64E()
	return
}

// GetUint64 returns input value as uint64 or the default value while it's present and input is blank.
func (v Values) Uint64E() (uint64, error) {
	if f, ok := v.first(); ok {
		return f.Uint64E()
	}
	return 0, errors.New("does not exist")
}

func (v Values) Bool() (i bool) {
	i, _ = v.BoolE()
	return
}

// GetBool returns input value as bool or the default value while it's present and input is blank.
func (v Values) BoolE() (bool, error) {
	if f, ok := v.first(); ok {
		return f.BoolE()
	}
	return false, errors.New("does not exist")
}

func (v Values) Float() (i float64) {
	i, _ = v.FloatE()
	return
}

// GetFloat returns input value as float64 or the default value while it's present and input is blank.
func (v Values) FloatE() (float64, error) {
	if f, ok := v.first(); ok {
		return f.FloatE()
	}
	return 0, errors.New("does not exist")
}

type Value string

// GetInt returns input as an int or the default value while it's present and input is blank
func (v Value) Int() int {
	i, _ := v.IntE()
	return i
}

func (v Value) IntE() (int, error) {
	return strconv.Atoi(string(v))
}

func (v Value) Int8() int8 {
	i, _ := v.Int8E()
	return i
}

// Int8E return input as an int8 or the default value while it's present and input is blank
func (v Value) Int8E() (int8, error) {
	i64, err := strconv.ParseInt(string(v), 10, 8)
	return int8(i64), err
}

// GetUint8 return input as an uint8 or the default value while it's present and input is blank
func (v Value) Uint8() (i uint8) {
	i, _ = v.Uint8E()
	return
}

// GetUint8 return input as an uint8 or the default value while it's present and input is blank
func (v Value) Uint8E() (uint8, error) {
	u64, err := strconv.ParseUint(string(v), 10, 8)
	return uint8(u64), err
}

func (v Value) Int16() (i int16) {
	i, _ = v.Int16E()
	return
}

// GetInt16 returns input as an int16 or the default value while it's present and input is blank
func (v Value) Int16E() (int16, error) {
	i64, err := strconv.ParseInt(string(v), 10, 16)
	return int16(i64), err
}

func (v Value) Uint16() (i uint16) {
	i, _ = v.Uint16E()
	return
}

// GetUint16 returns input as an uint16 or the default value while it's present and input is blank
func (v Value) Uint16E() (uint16, error) {
	u64, err := strconv.ParseUint(string(v), 10, 16)
	return uint16(u64), err
}

func (v Value) Int32() (i int32) {
	i, _ = v.Int32E()
	return
}

// GetInt32 returns input as an int32 or the default value while it's present and input is blank
func (v Value) Int32E() (int32, error) {
	i64, err := strconv.ParseInt(string(v), 10, 32)
	return int32(i64), err
}

func (v Value) Uint32() (i uint32) {
	i, _ = v.Uint32E()
	return
}

// GetUint32 returns input as an uint32 or the default value while it's present and input is blank
func (v Value) Uint32E() (uint32, error) {
	u64, err := strconv.ParseUint(string(v), 10, 32)
	return uint32(u64), err
}

func (v Value) Int64() (i int64) {
	i, _ = v.Int64E()
	return
}

// GetInt64 returns input value as int64 or the default value while it's present and input is blank.
func (v Value) Int64E() (int64, error) {
	return strconv.ParseInt(string(v), 10, 64)
}

func (v Value) Uint64() (i uint64) {
	i, _ = v.Uint64E()
	return
}

// GetUint64 returns input value as uint64 or the default value while it's present and input is blank.
func (v Value) Uint64E() (uint64, error) {
	return strconv.ParseUint(string(v), 10, 64)
}

func (v Value) Bool() (i bool) {
	i, _ = v.BoolE()
	return
}

// GetBool returns input value as bool or the default value while it's present and input is blank.
func (v Value) BoolE() (bool, error) {
	return strconv.ParseBool(string(v))
}

func (v Value) Float() (i float64) {
	i, _ = v.FloatE()
	return
}

// GetFloat returns input value as float64 or the default value while it's present and input is blank.
func (v Value) FloatE() (float64, error) {
	return strconv.ParseFloat(string(v), 64)
}

// GetFloat returns input value as float64 or the default value while it's present and input is blank.
func (v Value) String() string {
	return string(v)
}
