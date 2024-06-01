package schema

import (
	"bytes"
	"encoding/json"
)

type Properties []Property

func (p Properties) MarshalJSON() ([]byte, error) {
	visited := make(map[string]bool)
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, prop := range p {
		if visited[prop.Name] {
			continue
		}
		visited[prop.Name] = true

		if i > 0 {
			buf.WriteString(",")
		}
		if err := json.NewEncoder(&buf).Encode(prop.Name); err != nil {
			return nil, err
		}
		buf.WriteString(`:`)
		if err := json.NewEncoder(&buf).Encode(prop.Definition); err != nil {
			return nil, err
		}
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

type Property struct {
	Name string `json:"-"`
	Definition
}
