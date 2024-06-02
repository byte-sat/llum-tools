package tools

import (
	"fmt"
	"reflect"
)

type Injector struct {
	parent *Injector
	vals   map[reflect.Type]func() (any, error)
}

func Inject(vals ...any) (*Injector, error) {
	inj := &Injector{
		vals: make(map[reflect.Type]func() (any, error)),
	}

	for _, val := range vals {
		if err := inj.Provide(val); err != nil {
			return nil, err
		}
	}

	return inj, nil
}

func (i *Injector) Provide(val any) error {
	rv := reflect.ValueOf(val)
	t := rv.Type()

	if rv.Kind() != reflect.Func {
		i.vals[t] = func() (any, error) { return val, nil }
		return nil
	}

	if t.NumIn() != 0 {
		return fmt.Errorf("function must have no arguments")
	}

	if t.NumOut() == 2 {
		i.vals[t] = val.(func() (any, error))
		return nil
	}

	if t.NumOut() != 1 {
		return fmt.Errorf("function must have one or two return values")
	}

	out := t.Out(0)
	i.vals[out] = func() (any, error) {
		outs := rv.Call(nil)
		return outs[0].Interface(), nil
	}

	return nil
}

func (i *Injector) has(t reflect.Type) bool {
	if _, ok := i.vals[t]; ok {
		return true
	}

	if i.parent != nil {
		return i.parent.has(t)
	}

	return false
}

func (i *Injector) get(t reflect.Type) (any, error) {
	if fn, ok := i.vals[t]; ok {
		return fn()
	}

	if i.parent != nil {
		return i.parent.get(t)
	}

	return nil, fmt.Errorf("no provider for type: %s", t)
}
