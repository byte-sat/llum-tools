package tools

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/byte-sat/llum-tools/schema"
)

func Func(fn any, desc string, args ...string) Tool {
	if len(args)%2 != 0 {
		panic("all arguments must have a name and description")
	}

	rfn := reflect.ValueOf(fn)
	if rfn.Kind() != reflect.Func {
		panic("fn must be a function")
	}

	fnt := rfn.Type()
	_, name, _ := strings.Cut(runtime.FuncForPC(rfn.Pointer()).Name(), ".")

	if fnt.NumOut() > 2 {
		panic("function has too many outs")
	}

	props := make([]schema.Property, 0, len(args)/2)
	names := make([]string, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		props = append(props, schema.Property{
			Name: args[i],
			Definition: schema.Definition{
				Description: args[i+1],
			},
		})
		names = append(names, args[i])
	}

	var specials int
	ins := make([]reflect.Type, 0, fnt.NumIn())
	for i := 0; i < fnt.NumIn(); i++ {
		it := fnt.In(i)

		if it.Kind() == reflect.Interface && it.Implements(ctxType) {
			specials++
		}

		ins = append(ins, it)

		var propType schema.Type
		switch it.Kind() {
		case reflect.Bool:
			propType = schema.Boolean
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			propType = schema.Integer
		case reflect.Float32, reflect.Float64:
			propType = schema.Number
		case reflect.String:
			propType = schema.String
		default:
			panic(fmt.Sprintf("unsupported argument type %s", it.String()))

		}

		props[i-specials].Type = propType
	}

	if len(props)+specials != len(ins) {
		panic("names must have the same length as the function's arguments")
	}

	outs := make([]reflect.Type, fnt.NumOut())
	for i := range outs {
		outs[i] = fnt.Out(i)
	}

	if len(outs) == 2 {
		if !outs[1].AssignableTo(reflect.TypeFor[error]()) {
			panic("second return value must be an error")
		}
	}

	inv := &funcInvoker{
		rv:       rfn,
		ins:      ins,
		names:    names,
		specials: specials,
	}

	return Tool{
		Function: schema.Function{
			Name:        name,
			Description: desc,
			Parameters: schema.Definition{
				Type:       schema.Object,
				Properties: props,
				Required:   names,
			},
		},
		Invoker: inv,
	}
}

type funcInvoker struct {
	rv       reflect.Value
	ins      []reflect.Type
	names    []string
	specials int
}

func (f *funcInvoker) Invoke(context context.Context, args map[string]any) (r any, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r)
			log.Println(string(debug.Stack()))
			err = errors.New("something bad happened")
		}
	}()
	nameIds := make(map[string]int, len(f.names))
	for i, name := range f.names {
		nameIds[name] = f.specials + i
		if _, ok := args[name]; !ok {
			return nil, fmt.Errorf("missing argument: %s", name)
		}
	}

	ins := make([]reflect.Value, len(f.ins))
	for i := 0; i < f.specials; i++ {
		it := f.ins[i]
		if it.Kind() == reflect.Interface && it.Implements(ctxType) {
			ins[i] = reflect.ValueOf(context)
		}
	}

	for name, value := range args {
		id, ok := nameIds[name]
		if !ok {
			return nil, fmt.Errorf("unexpected argument: %s", name)
		}

		ins[id], err = convert(f.ins[id], value)
		if err != nil {
			return nil, fmt.Errorf("argument %s: %w", name, err)
		}
	}

	outs := f.rv.Call(ins)
	if len(outs) == 0 {
		return nil, nil
	}

	if len(outs) == 2 {
		out2 := outs[1]
		if !out2.IsNil() {
			return nil, out2.Interface().(error)
		}
	}

	out := outs[0]
	if err, ok := out.Interface().(error); ok && err != nil {
		return nil, err
	}

	return out.Interface(), nil
}

func convert(t reflect.Type, val any) (reflect.Value, error) {
	rv := reflect.ValueOf(val)
	if rv.Type().ConvertibleTo(t) {
		return rv.Convert(t), nil
	}

	switch v := val.(type) {
	case string:
		switch t.Kind() {
		case reflect.Bool:
			val, err := strconv.ParseBool(v)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(val), nil

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(num).Convert(t), nil

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			num, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(num).Convert(t), nil

		case reflect.Float32, reflect.Float64:
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(num).Convert(t), nil
		}

		val := reflect.Zero(t)
		if tm, ok := val.Interface().(encoding.TextUnmarshaler); ok {
			if err := tm.UnmarshalText([]byte(v)); err != nil {
				return reflect.Value{}, err
			}
			return val, nil
		}

	}

	return rv, nil
}
