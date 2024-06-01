package schema

import "encoding/json"

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Definition `json:"parameters,omitempty"`
}

func (f Function) MarshalJSON() ([]byte, error) {
	type alias Function
	tool := struct {
		Type     string `json:"type"`
		Function alias  `json:"function"`
	}{Type: "function", Function: alias(f)}
	return json.Marshal(tool)
}
