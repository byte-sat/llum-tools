package tools

import (
	"context"
	"fmt"

	"github.com/byte-sat/llum-tools/schema"
)

type Repo struct {
	tools  map[string]Invoker
	schema []schema.Function
}

func New(fns ...any) (*Repo, error) {
	r := &Repo{
		tools:  make(map[string]Invoker, len(fns)),
		schema: make([]schema.Function, 0, len(fns)),
	}

	for _, fn := range fns {
		if err := r.Add(fn); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repo) Add(fn any) error {
	schema, invoker := codocFunc(fn)
	r.tools[schema.Name] = invoker
	r.schema = append(r.schema, schema)
	return nil
}

func (r *Repo) Schema() []schema.Function { return r.schema }

func (r *Repo) Invoke(ctx context.Context, name string, args map[string]any) (any, error) {
	if invoker, ok := r.tools[name]; ok {
		return invoker.Invoke(ctx, args)
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}
