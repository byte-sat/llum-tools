package tools

import (
	"fmt"
	"reflect"
	"strconv"
)

type argConverter func(any) (reflect.Value, error)

func converter(t reflect.Type) argConverter {
	switch t.Kind() {
	case reflect.Bool:
		return parseBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return parseInt(t)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parseUint(t)
	case reflect.Float32, reflect.Float64:
		return parseFloat(t)
	case reflect.String:
		return parseNotImpl("string")
	case reflect.Array:
		return parseNotImpl("array")
	case reflect.Slice:
		return parseNotImpl("slice")
	case reflect.Struct:
		return parseNotImpl("struct")
	case reflect.Map:
		return parseNotImpl("map")
	default:
		panic(fmt.Sprintf("unsupported argument type %s", t.Kind()))

	}
}

func parseBool(val any) (reflect.Value, error) {
	switch v := val.(type) {
	case string:
		var err error
		val, err = strconv.ParseBool(v)
		if err != nil {
			return reflect.Value{}, err
		}
	case int:
		val = v != 0
	case int8:
		val = v != 0
	case int16:
		val = v != 0
	case int32:
		val = v != 0
	case int64:
		val = v != 0
	case uint8:
		val = v != 0
	case uint16:
		val = v != 0
	case uint32:
		val = v != 0
	case uint64:
		val = v != 0
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", val)
	}

	return reflect.ValueOf(val), nil
}

func parseInt(t reflect.Type) argConverter {
	return func(val any) (reflect.Value, error) {
		switch v := val.(type) {
		case string:
			var err error
			val, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}

		case float32:
			val = int(v)
		case float64:
			val = int(v)

		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		return reflect.ValueOf(val).Convert(t), nil
	}
}

func parseUint(t reflect.Type) argConverter {
	return func(val any) (reflect.Value, error) {
		switch v := val.(type) {
		case string:
			var err error
			val, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}

		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		return reflect.ValueOf(val).Convert(t), nil
	}
}

func parseFloat(t reflect.Type) argConverter {
	return func(val any) (reflect.Value, error) {
		switch v := val.(type) {
		case string:
			var err error
			val, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return reflect.Value{}, err
			}

		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		return reflect.ValueOf(val).Convert(t), nil
	}
}

func parseNotImpl(kind string) argConverter {
	return func(val any) (reflect.Value, error) {
		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, kind)
	}
}
