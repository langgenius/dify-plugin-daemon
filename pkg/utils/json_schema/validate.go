package jsonschema

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateBytes is a low-level helper that validates a JSON document (bytes) against a JSON schema (bytes).
// It avoids passing live Go maps/slices to gojsonschema, preventing concurrent map iteration panics.
func ValidateBytes(schemaJSON []byte, docJSON []byte) (*gojsonschema.Result, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	docLoader := gojsonschema.NewBytesLoader(docJSON)
	return gojsonschema.Validate(schemaLoader, docLoader)
}

// Validate clones schema and document via json.Marshal and validates them.
// Use this when you might have live maps mutated by other goroutines.
func Validate(schema any, document any) (*gojsonschema.Result, error) {
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	docJSON, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	return ValidateBytes(schemaJSON, docJSON)
}
