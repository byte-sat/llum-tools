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
		return justConvert(t)
	case reflect.Array:
		return convertArray(t)
	case reflect.Slice:
		return convertSlice(t)
	case reflect.Struct:
		return convertStruct(t)
	case reflect.Map:
		return convertMap(t)
	case reflect.Ptr:
		return convertPtr(t)
	default:
		panic(fmt.Sprintf("unsupported argument type %s", t.Kind()))

	}
}

func convertPtr(t reflect.Type) argConverter {
	it := t.Elem()
	conv := converter(it)
	return func(val any) (reflect.Value, error) {
		rv, err := conv(val)
		if err != nil {
			return rv, err
		}

		ptr := reflect.New(it)
		ptr.Elem().Set(rv)

		return ptr, nil
	}
}

func convertArray(t reflect.Type) argConverter {
	conv := converter(t.Elem())
	return func(val any) (reflect.Value, error) {
		rv := reflect.ValueOf(val)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		if rv, err, ok := tryConvert(t, rv); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}

		ptr := reflect.New(t)
		arr := ptr.Elem()

		if rv.Len() != arr.Len() {
			return reflect.Value{}, fmt.Errorf("array length mismatch: %d != %d", rv.Len(), arr.Len())
		}

		for i := 0; i < rv.Len(); i++ {
			v := rv.Index(i)
			v, err := conv(v.Interface())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s[%d]: %w", t.String(), i, err)
			}

			arr.Index(i).Set(v)
		}

		return arr, nil
	}
}

func convertSlice(t reflect.Type) argConverter {
	conv := converter(t.Elem())
	return func(val any) (reflect.Value, error) {
		rv := reflect.ValueOf(val)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		if rv, err, ok := tryConvert(t, rv); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}

		ptr := reflect.New(t)
		sl := ptr.Elem()
		sl.Grow(rv.Len())
		sl.SetLen(rv.Len())

		for i := 0; i < rv.Len(); i++ {
			v := rv.Index(i)
			v, err := conv(v.Interface())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s[%d]: %w", t.String(), i, err)
			}

			sl.Index(i).Set(v)
		}

		return sl, nil
	}
}

func convertMap(t reflect.Type) argConverter {
	if t.Key().Kind() != reflect.String {
		panic("map keys must be strings")
	}

	conv := converter(t.Elem())
	return func(val any) (reflect.Value, error) {
		rv := reflect.ValueOf(val)
		if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		if rv, err, ok := tryConvert(t, rv); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}

		mv := reflect.MakeMapWithSize(t, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()

			v, err := conv(v.Interface())
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s[%s]: %w", t.String(), k.Interface(), err)
			}

			mv.SetMapIndex(k, v)
		}

		return mv, nil
	}
}

func convertStruct(t reflect.Type) argConverter {
	nameidx := map[string]int{}
	converters := make([]argConverter, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := fieldName(f)
		if name == "" {
			continue
		}

		nameidx[name] = i
		converters[i] = converter(f.Type)
	}

	return func(val any) (reflect.Value, error) {
		if rv, err, ok := tryConvert(t, reflect.ValueOf(val)); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}

		m, ok := val.(map[string]any)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
		}

		ptr := reflect.New(t)
		rv := ptr.Elem()
		for key, val := range m {
			idx, ok := nameidx[key]
			if !ok {
				return reflect.Value{}, fmt.Errorf("unexpected field: %s", key)
			}

			v, err := converters[idx](val)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%s.%s: %w", t.String(), t.Field(idx).Name, err)
			}

			rv.Field(idx).Set(v)
		}

		return rv, nil
	}
}

func parseInt(t reflect.Type) argConverter {
	return func(val any) (reflect.Value, error) {
		if rv, err, ok := tryConvert(t, reflect.ValueOf(val)); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}

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
		if rv, err, ok := tryConvert(t, reflect.ValueOf(val)); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}
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
		if rv, err, ok := tryConvert(t, reflect.ValueOf(val)); ok {
			if err != nil {
				return reflect.Value{}, err
			}
			if ok {
				return rv, nil
			}
		}
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
	case bool:
		val = v
	default:
		return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", val)
	}

	return reflect.ValueOf(val), nil
}

func justConvert(t reflect.Type) argConverter {
	return func(val any) (reflect.Value, error) {
		if rv, err, ok := tryConvert(t, reflect.ValueOf(val)); ok {
			return rv, err
		}

		return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, t)
	}
}

func tryConvert(t reflect.Type, rv reflect.Value) (reflect.Value, error, bool) {
	if t.AssignableTo(rv.Type()) {
		return rv, nil, true
	}

	if rv.CanConvert(t) {
		return rv.Convert(t), nil, true
	}

	return reflect.Value{}, nil, false
}
