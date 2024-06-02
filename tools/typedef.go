package tools

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/byte-sat/llum-tools/schema"
	"github.com/noonien/codoc"
)

func typeDefinition(t reflect.Type) schema.Definition {
	switch t.Kind() {
	case reflect.Pointer:
		return typeDefinition(t.Elem())

	case reflect.Bool:
		return schema.Definition{
			Type: schema.Boolean,
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return schema.Definition{
			Type: schema.Integer,
		}

	case reflect.Float32, reflect.Float64:
		return schema.Definition{
			Type: schema.Number,
		}

	case reflect.String:
		return schema.Definition{
			Type: schema.String,
		}

	case reflect.Struct:
		cs := codoc.GetStruct(t.String())
		props := make([]schema.Property, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if !f.IsExported() {
				continue
			}

			name := fieldName(f)
			if name == "" {
				continue
			}

			prop := schema.Property{
				Name:       name,
				Definition: typeDefinition(f.Type),
			}

			if cs != nil {
				f := cs.Fields[f.Name]
				if doc := f.Doc; doc != "" {
					prop.Description = doc
				} else {
					prop.Description = f.Comment
				}
			}

			props = append(props, prop)
		}

		return schema.Definition{
			Type:       schema.Object,
			Properties: props,
		}

	case reflect.Array, reflect.Slice:
		td := typeDefinition(t.Elem())
		return schema.Definition{
			Type:  schema.Array,
			Items: &td,
		}

	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic("map keys must be strings")
		}

		if t.Elem().Kind() == reflect.Interface {
			panic("map values cannot be interfaces")
		}

		td := typeDefinition(t.Elem())
		return schema.Definition{
			Type:  schema.Object,
			Items: &td,
		}

	default:
		panic(fmt.Sprintf("unsupported argument type %s", t.Kind()))
	}

}

func fieldName(f reflect.StructField) string {
	var name string
	if tag := f.Tag.Get("llm"); tag != "" {
		name, _, _ = strings.Cut(tag, ",")
	} else if tag := f.Tag.Get("json"); tag != "" {
		name, _, _ = strings.Cut(tag, ",")
	}
	if name == "-" {
		return ""
	}
	if name == "" {
		return f.Name
	}
	return name
}
