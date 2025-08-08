package jsonschema

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/stretchr/testify/assert"
)

func TestGenerateValidateJson(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
		},
	}

	type nameStruct struct {
		Name string `json:"name"`
	}

	validateJson, err := GenerateValidateJson(schema)
	assert.NoError(t, err)

	// convert validateJson to nameStruct
	name, err := parser.MapToStruct[nameStruct](validateJson)
	assert.NoError(t, err)

	assert.NotEmpty(t, name.Name)
}

func TestGenerateValidateJsonWithEnum(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
				"enum": []string{"test", "test2"},
			},
		},
	}

	type nameStruct struct {
		Name string `json:"name"`
	}

	validateJson, err := GenerateValidateJson(schema)
	assert.NoError(t, err)

	// convert validateJson to nameStruct
	name, err := parser.MapToStruct[nameStruct](validateJson)
	assert.NoError(t, err)

	assert.Contains(t, []string{"test", "test2"}, name.Name)
}

func TestGenerateValidateJsonWithNumber(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"age": map[string]any{
				"type": "integer",
			},
		},
	}

	type nameStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	validateJson, err := GenerateValidateJson(schema)
	assert.NoError(t, err)

	// convert validateJson to nameStruct
	name, err := parser.MapToStruct[nameStruct](validateJson)
	assert.NoError(t, err)

	assert.NotEmpty(t, name.Name)
	assert.NotEmpty(t, name.Age)
}
