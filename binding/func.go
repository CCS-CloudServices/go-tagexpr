package binding

import (
	jsonpkg "encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/henrylee2cn/ameda"
)

// JSONUnmarshaler is the interface implemented by types
// that can unmarshal a JSON description of themselves.
type JSONUnmarshaler func(data []byte, v interface{}) error

var (
	jsonUnmarshalFunc func(data []byte, v interface{}) error

	kindTypeMap = map[reflect.Kind]reflect.Type{
		reflect.String:  reflect.TypeOf(""),
		reflect.Bool:    reflect.TypeOf(false),
		reflect.Float32: reflect.TypeOf(float32(0)),
		reflect.Float64: reflect.TypeOf(float64(0)),
		reflect.Int:     reflect.TypeOf(int(0)),
		reflect.Int64:   reflect.TypeOf(int64(0)),
		reflect.Int32:   reflect.TypeOf(int32(0)),
		reflect.Int16:   reflect.TypeOf(int16(0)),
		reflect.Int8:    reflect.TypeOf(int8(0)),
		reflect.Uint:    reflect.TypeOf(uint(0)),
		reflect.Uint64:  reflect.TypeOf(uint64(0)),
		reflect.Uint32:  reflect.TypeOf(uint32(0)),
		reflect.Uint16:  reflect.TypeOf(uint16(0)),
		reflect.Uint8:   reflect.TypeOf(uint8(0)),
	}
)

// ResetJSONUnmarshaler reset the JSON Unmarshal function.
// NOTE: verifyingRequired is true if the required tag is supported.
func ResetJSONUnmarshaler(fn JSONUnmarshaler) {
	jsonUnmarshalFunc = fn
}

var typeUnmarshalFuncs = make(map[reflect.Type]func(string, bool) (reflect.Value, error))

func unsafeUnmarshalValue(v reflect.Value, s string, looseZeroMode bool) error {
	fn := typeUnmarshalFuncs[v.Type()]
	if fn != nil {
		vv, err := fn(s, looseZeroMode)
		if err == nil {
			v.Set(vv)
		}
		return err
	}
	return unmarshal(ameda.UnsafeStringToBytes(s), v.Addr().Interface())
}

func unmarshalSlice(fn func(string, bool) (reflect.Value, error), t reflect.Type, a []string, looseZeroMode bool) (
	reflect.Value, error) {
	var err error
	v := reflect.New(reflect.SliceOf(t)).Elem()
	for _, s := range a {
		var vv reflect.Value
		vv, err = fn(s, looseZeroMode)
		if err != nil {
			return v, err
		}
		v = reflect.Append(v, vv)
	}
	return v, nil
}

func unmarshal(b []byte, i interface{}) error {
	switch x := i.(type) {
	case jsonpkg.Unmarshaler:
		return x.UnmarshalJSON(b)
	case proto.Unmarshaler:
		return x.Unmarshal(b)
	default:
		return jsonpkg.Unmarshal(b, i)
	}
}

// MustRegTypeUnmarshal registers unmarshalor function of type.
// NOTE:
//  panic if exist error.
func MustRegTypeUnmarshal(t reflect.Type, fn func(v string, emptyAsZero bool) (reflect.Value, error)) {
	err := RegTypeUnmarshal(t, fn)
	if err != nil {
		panic(err)
	}
}

// RegTypeUnmarshal registers unmarshalor function of type.
func RegTypeUnmarshal(t reflect.Type, fn func(v string, emptyAsZero bool) (reflect.Value, error)) error {
	// check
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
		reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		otherType, ok := kindTypeMap[t.Kind()]
		if !ok {
			return errors.New("basic type not supported in map")
		}
		if t == otherType {
			return errors.New("registration type cannot be a basic type")
		}
	case reflect.Ptr:
		return errors.New("registration type cannot be a pointer type")
	}
	// test
	vv, err := fn("", true)
	if err != nil {
		return fmt.Errorf("test fail: %s", err)
	}
	if tt := vv.Type(); tt != t {
		return fmt.Errorf("test fail: expect return value type is %s, but got %s", t.String(), tt.String())
	}

	typeUnmarshalFuncs[t] = fn
	return nil
}

func init() {
	MustRegTypeUnmarshal(reflect.TypeOf(time.Time{}), func(v string, emptyAsZero bool) (reflect.Value, error) {
		if v == "" && emptyAsZero {
			return reflect.ValueOf(time.Time{}), nil
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(t), nil
	})
}
