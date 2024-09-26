package hi

import (
	"errors"
	"strconv"
)

// todo: 待整理

// func (c *Context) Query2(key string, defaultValue ...string) (value to.Value) {
// 	c.initQueryCache()
// 	values, ok := c.queryCache[key]
// 	if !ok || len(values) == 0 {
// 		num := len(defaultValue)
// 		if num == 0 {
// 			return
// 		}
// 		return to.ValueF(defaultValue[0])
// 	}
// 	return to.ValueF(values[0])
// }

// func (c *Context) Query2Array(key string, defaultValue ...any) (nValues to.Values[string]) {
// 	c.initQueryCache()
// 	values, ok := c.queryCache[key]
// 	if ok {
// 		return to.ValuesF(values)
// 	}
// 	if len(defaultValue) == 0 {
// 		return
// 	}
// 	return to.ValuesF(to.ValuesF(defaultValue).String())
// }

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

///////////////////////////////////////////////////////////////////////

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key any) any {
	if key == ContextRequestKey {
		return c.Request
	}
	if key == ContextKey {
		return c
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Value(key)
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// // MustGet returns the value for the given key if it exists, otherwise it panics.
// func (c *Context) MustGet(key string) any {
// 	if value, exists := c.Get(key); exists {
// 		return value
// 	}
// 	panic("Key \"" + key + "\" does not exist")
// }

// // GetString returns the value associated with the key as a string.
// func (c *Context) GetString(key string) (s string) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		s, _ = val.(string)
// 	}
// 	return
// }

// // GetBool returns the value associated with the key as a boolean.
// func (c *Context) GetBool(key string) (b bool) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		b, _ = val.(bool)
// 	}
// 	return
// }

// // GetInt returns the value associated with the key as an integer.
// func (c *Context) GetInt(key string) (i int) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i, _ = val.(int)
// 	}
// 	return
// }

// // GetInt8 returns the value associated with the key as an integer 8.
// func (c *Context) GetInt8(key string) (i8 int8) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i8, _ = val.(int8)
// 	}
// 	return
// }

// // GetInt16 returns the value associated with the key as an integer 16.
// func (c *Context) GetInt16(key string) (i16 int16) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i16, _ = val.(int16)
// 	}
// 	return
// }

// // GetInt32 returns the value associated with the key as an integer 32.
// func (c *Context) GetInt32(key string) (i32 int32) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i32, _ = val.(int32)
// 	}
// 	return
// }

// // GetInt64 returns the value associated with the key as an integer 64.
// func (c *Context) GetInt64(key string) (i64 int64) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i64, _ = val.(int64)
// 	}
// 	return
// }

// // GetUint returns the value associated with the key as an unsigned integer.
// func (c *Context) GetUint(key string) (ui uint) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui, _ = val.(uint)
// 	}
// 	return
// }

// // GetUint8 returns the value associated with the key as an unsigned integer 8.
// func (c *Context) GetUint8(key string) (ui8 uint8) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui8, _ = val.(uint8)
// 	}
// 	return
// }

// // GetUint16 returns the value associated with the key as an unsigned integer 16.
// func (c *Context) GetUint16(key string) (ui16 uint16) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui16, _ = val.(uint16)
// 	}
// 	return
// }

// // GetUint32 returns the value associated with the key as an unsigned integer 32.
// func (c *Context) GetUint32(key string) (ui32 uint32) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui32, _ = val.(uint32)
// 	}
// 	return
// }

// // GetUint64 returns the value associated with the key as an unsigned integer 64.
// func (c *Context) GetUint64(key string) (ui64 uint64) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui64, _ = val.(uint64)
// 	}
// 	return
// }

// // GetFloat32 returns the value associated with the key as a float32.
// func (c *Context) GetFloat32(key string) (f32 float32) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		f32, _ = val.(float32)
// 	}
// 	return
// }

// // GetFloat64 returns the value associated with the key as a float64.
// func (c *Context) GetFloat64(key string) (f64 float64) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		f64, _ = val.(float64)
// 	}
// 	return
// }

// // GetTime returns the value associated with the key as time.
// func (c *Context) GetTime(key string) (t time.Time) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		t, _ = val.(time.Time)
// 	}
// 	return
// }

// // GetDuration returns the value associated with the key as a duration.
// func (c *Context) GetDuration(key string) (d time.Duration) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		d, _ = val.(time.Duration)
// 	}
// 	return
// }

// func (c *Context) GetIntSlice(key string) (is []int) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		is, _ = val.([]int)
// 	}
// 	return
// }

// func (c *Context) GetInt8Slice(key string) (i8s []int8) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i8s, _ = val.([]int8)
// 	}
// 	return
// }

// func (c *Context) GetInt16Slice(key string) (i16s []int16) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i16s, _ = val.([]int16)
// 	}
// 	return
// }

// func (c *Context) GetInt32Slice(key string) (i32s []int32) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i32s, _ = val.([]int32)
// 	}
// 	return
// }

// func (c *Context) GetInt64Slice(key string) (i64s []int64) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		i64s, _ = val.([]int64)
// 	}
// 	return
// }

// func (c *Context) GetUintSlice(key string) (uis []uint) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		uis, _ = val.([]uint)
// 	}
// 	return
// }

// func (c *Context) GetUint8Slice(key string) (ui8s []uint8) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui8s, _ = val.([]uint8)
// 	}
// 	return
// }

// func (c *Context) GetUint16Slice(key string) (ui16s []uint16) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui16s, _ = val.([]uint16)
// 	}
// 	return
// }

// func (c *Context) GetUint32Slice(key string) (ui32s []uint32) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui32s, _ = val.([]uint32)
// 	}
// 	return
// }

// func (c *Context) GetUint64Slice(key string) (ui64s []uint64) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ui64s, _ = val.([]uint64)
// 	}
// 	return
// }

// func (c *Context) GetFloat32Slice(key string) (f32s []float32) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		f32s, _ = val.([]float32)
// 	}
// 	return
// }

// func (c *Context) GetFloat64Slice(key string) (f64s []float64) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		f64s, _ = val.([]float64)
// 	}
// 	return
// }

// // GetStringSlice returns the value associated with the key as a slice of strings.
// func (c *Context) GetStringSlice(key string) (ss []string) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		ss, _ = val.([]string)
// 	}
// 	return
// }

// // GetStringMap returns the value associated with the key as a map of interfaces.
// func (c *Context) GetStringMap(key string) (sm map[string]any) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		sm, _ = val.(map[string]any)
// 	}
// 	return
// }

// // GetStringMapString returns the value associated with the key as a map of strings.
// func (c *Context) GetStringMapString(key string) (sms map[string]string) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		sms, _ = val.(map[string]string)
// 	}
// 	return
// }

// // GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
// func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
// 	if val, ok := c.Get(key); ok && val != nil {
// 		smss, _ = val.(map[string][]string)
// 	}
// 	return
// }
