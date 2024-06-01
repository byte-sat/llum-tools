package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/byte-sat/llum-tools/schema"
)

type Group struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SubGroups   []Group `json:"subgroups,omitempty"`
	Tools       []Tool  `json:"tools,omitempty"`
}

type Tool struct {
	schema.Function
	Invoker Invoker `json:"-"`
}

func (g Group) MarshalJSON() ([]byte, error) {
	tools := []schema.Function{}
	tools = g.appendTools(tools, "")

	return json.Marshal(tools)
}

func (g *Group) appendTools(tools []schema.Function, prefix string) []schema.Function {
	if prefix != "" {
		prefix += "."
	}

	for _, tool := range g.Tools {
		tool.Function.Name = prefix + tool.Function.Name
		tools = append(tools, tool.Function)
	}

	for _, group := range g.SubGroups {
		tools = group.appendTools(tools, prefix+group.Name)
	}

	return tools
}

func (g *Group) Invoke(ctx context.Context, name string, args map[string]any) (any, error) {
	for _, tool := range g.Tools {
		if tool.Name == name {
			return tool.Invoker.Invoke(ctx, args)
		}
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}
