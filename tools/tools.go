package tools

import (
	"fmt"

	"github.com/byte-sat/llum-tools/schema"
)

type Repo struct {
	tools  map[string]invoker
	schema []schema.Function
	inj    *Injector
}

type invoker interface {
	Invoke(*Injector, map[string]any) (any, error)
}

func New(inj *Injector, fns ...any) (*Repo, error) {
	if inj == nil {
		inj, _ = Inject()
	}
	r := &Repo{
		tools:  make(map[string]invoker, len(fns)),
		schema: make([]schema.Function, 0, len(fns)),
		inj:    inj,
	}

	for _, fn := range fns {
		if err := r.Add(fn); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repo) Add(fn any) error {
	schema, invoker := codocFunc(r.inj, fn)
	r.tools[schema.Name] = invoker
	r.schema = append(r.schema, schema)
	return nil
}

func (r *Repo) Schema() []schema.Function { return r.schema }

func (r *Repo) Invoke(inj *Injector, name string, args map[string]any) (any, error) {
	if invoker, ok := r.tools[name]; ok {
		return invoker.Invoke(inj, args)
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}
