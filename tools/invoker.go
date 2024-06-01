package tools

import "context"

type Invoker interface {
	Invoke(context.Context, map[string]any) (any, error)
}
