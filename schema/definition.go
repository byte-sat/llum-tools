package schema

type Type string

const (
	Object  Type = "object"
	Number  Type = "number"
	Integer Type = "integer"
	String  Type = "string"
	Array   Type = "array"
	Null    Type = "null"
	Boolean Type = "boolean"
)

// Definition is a struct for describing a JSON Schema.
type Definition struct {
	// Type specifies the data type of the schema.
	Type Type `json:"type,omitempty"`
	// Description is the description of the schema.
	Description string `json:"description,omitempty"`
	// Enum is used to restrict a value to a fixed set of values. It must be an array with at least
	// one element, where each element is unique. You will probably only use this with strings.
	Enum []string `json:"enum,omitempty"`
	// Properties describes the properties of an object, if the schema type is Object.
	Properties Properties `json:"properties,omitempty"`
	// Required specifies which properties are required, if the schema type is Object.
	Required []string `json:"required,omitempty"`
	// Items specifies which data type an array contains, if the schema type is Array.
	Items *Definition `json:"items,omitempty"`
}
