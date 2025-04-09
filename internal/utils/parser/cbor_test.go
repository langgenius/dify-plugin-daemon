package parser

import (
	"testing"
)

func TestMarshalCBORMap(t *testing.T) {
	data := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
	}

	tool, err := MarshalCBOR(data)
	if err != nil {
		t.Fatal(err)
	}

	m, err := UnmarshalCBOR[map[string]any](tool)
	if err != nil {
		t.Fatal(err)
	}

	// check properties is map[string]any
	if _, ok := m["properties"]; !ok {
		t.Fatal("properties dose not exist")
	}

	_, ok := m["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties is not a map[string]any")
	}
}
