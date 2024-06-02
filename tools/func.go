package tools

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/byte-sat/llum-tools/schema"
	"github.com/noonien/codoc"
)

var argRegex = regexp.MustCompile(`(?m)^([a-zA-Z_][a-zA-Z0-9_]*): (.+)$`)

var errType = reflect.TypeFor[error]()

func codocFunc(inj *Injector, fn any) (schema.Function, invoker) {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() != reflect.Func {
		panic("fn must be a function")
	}
	fnt := fnv.Type()

	if fnt.NumOut() > 2 {
		panic("function has too many outs")
	}

	if fnt.IsVariadic() {
		panic("variadic functions not supported")
	}

	ptr := fnv.Pointer()
	fullName := runtime.FuncForPC(ptr).Name()
	_, name, _ := strings.Cut(fullName, ".")

	doc := codoc.GetFunction(fullName)
	if doc == nil {
		panic("missing codoc documentation")
	}

	argDescs := make(map[string]string)
	for _, match := range argRegex.FindAllStringSubmatch(doc.Doc, -1) {
		arg, desc := match[1], match[2]
		argDescs[arg] = desc
	}

	desc := doc.Doc
	if len(argDescs) > 0 {
		idx := argRegex.FindStringIndex(doc.Doc)
		desc = strings.TrimSpace(doc.Doc[:idx[0]])
	}

	injected := make([]reflect.Type, 0, fnt.NumIn())
	for i := 0; i < fnt.NumIn(); i++ {
		it := fnt.In(i)
		if !inj.has(it) {
			break
		}

		injected = append(injected, it)
	}

	argNo := fnt.NumIn() - len(injected)
	argNames := make([]string, 0, argNo)
	argConverters := make([]argConverter, 0, argNo)
	args := make([]schema.Property, 0, argNo)

	for i := len(injected); i < fnt.NumIn(); i++ {
		it := fnt.In(i)
		name := doc.Args[i]

		def := typeDefinition(it)
		def.Description = argDescs[name]

		argNames = append(argNames, name)
		argConverters = append(argConverters, converter(it))
		args = append(args, schema.Property{
			Name:       name,
			Definition: def,
		})
	}

	var outfn func([]reflect.Value) (any, error)
	switch fnt.NumOut() {
	case 0:
		outfn = func([]reflect.Value) (any, error) { return nil, nil }
	case 1:
		if fnt.Out(0) == errType {
			outfn = func(outs []reflect.Value) (any, error) {
				return nil, outs[0].Interface().(error)
			}
		} else {
			outfn = func(outs []reflect.Value) (any, error) {
				return outs[0].Interface(), nil
			}
		}
	case 2:
		if fnt.Out(1) != errType {
			panic("second return value must be an error")
		}
		outfn = func(outs []reflect.Value) (any, error) {
			return outs[0].Interface(), outs[1].Interface().(error)
		}
	}

	fschema := schema.Function{
		Name:        name,
		Description: desc,
		Parameters: schema.Definition{
			Type:       schema.Object,
			Properties: args,
			Required:   argNames,
		},
	}

	if len(argNames) == 0 {
		fschema.Parameters = schema.Definition{}
	}

	return fschema, &codocFuncInvoker{
		injected:      injected,
		argConverters: argConverters,
		outfn:         outfn,
		argNames:      argNames,
		fnv:           fnv,
	}
}

type codocFuncInvoker struct {
	injected      []reflect.Type
	inj           *Injector
	argConverters []argConverter
	argNames      []string
	outfn         func([]reflect.Value) (any, error)
	fnv           reflect.Value
}

func (f *codocFuncInvoker) Invoke(inj *Injector, args map[string]any) (any, error) {
	if inj == nil {
		inj = f.inj
	} else {
		i := *inj
		i.parent = f.inj
		inj = &i
	}

	visited := map[string]bool{}
	vals := make([]reflect.Value, 0, len(f.argNames)+len(f.injected))
	for _, it := range f.injected {
		val, err := inj.get(it)
		if err != nil {
			return nil, err
		}
		vals = append(vals, reflect.ValueOf(val))
	}
	for i, name := range f.argNames {
		arg, ok := args[name]
		if !ok {
			return nil, fmt.Errorf("missing argument: %s", name)
		}
		val, err := f.argConverters[i](arg)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
		visited[name] = true
	}
	for name := range args {
		if !visited[name] {
			return nil, fmt.Errorf("unexpected argument: %s", name)
		}

	}

	outs := f.fnv.Call(vals)
	return f.outfn(outs)
}
