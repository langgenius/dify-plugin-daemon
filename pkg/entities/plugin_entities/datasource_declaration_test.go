package plugin_entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltinDefinitions(t *testing.T) {
	requiredTypes := []string{
		"file",
		"website_crawl",
		"online_document",
		"general_structure_chunk",
		"parent_child_structure_chunk",
		"qa_structure_chunk",
	}

	for _, typeName := range requiredTypes {
		assert.Contains(t, BuiltinDefinitions, typeName, "missing builtin type: %s", typeName)

		def := BuiltinDefinitions[typeName].(map[string]any)
		assert.Equal(t, "object", def["type"], "type %s should be object", typeName)
		assert.Contains(t, def, "properties", "type %s missing properties", typeName)

		props := def["properties"].(map[string]any)
		assert.Contains(t, props, "dify_builtin_type", "type %s missing dify_builtin_type", typeName)
	}
}

func TestResolveSchemaRefs(t *testing.T) {
	t.Run("resolve builtin type refs", func(t *testing.T) {
		schema := map[string]any{
			"$ref": "#/$defs/file",
		}

		resolved, err := ResolveSchemaRefs(schema, BuiltinDefinitions)
		assert.NoError(t, err)
		assert.Equal(t, "object", resolved.(map[string]any)["type"])
		assert.Contains(t, resolved.(map[string]any)["properties"].(map[string]any), "name")
	})

	t.Run("nested ref resolution", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"files": map[string]any{
					"type": "array",
					"items": map[string]any{
						"$ref": "#/$defs/file",
					},
				},
			},
		}

		resolved, err := ResolveSchemaRefs(schema, BuiltinDefinitions)
		assert.NoError(t, err)

		props := resolved.(map[string]any)["properties"].(map[string]any)
		items := props["files"].(map[string]any)["items"].(map[string]any)
		assert.Equal(t, "object", items["type"])
	})

	t.Run("non-existent ref returns error", func(t *testing.T) {
		schema := map[string]any{
			"$ref": "#/$defs/non_existent",
		}

		_, err := ResolveSchemaRefs(schema, BuiltinDefinitions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestProcessDatasourceYAML(t *testing.T) {
	t.Run("process output_schema with refs", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{
						"$ref": "#/$defs/file",
					},
				},
			},
		}

		processed, err := ProcessDatasourceYAML(yamlData)
		assert.NoError(t, err)

		outputSchema := processed["output_schema"].(map[string]any)
		result := outputSchema["properties"].(map[string]any)["result"].(map[string]any)
		assert.Equal(t, "object", result["type"])
		assert.Contains(t, result, "properties")
	})

	t.Run("process mixed user and builtin definitions", func(t *testing.T) {
		yamlData := map[string]any{
			"output_schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file": map[string]any{
						"$ref": "#/$defs/file",
					},
					"custom": map[string]any{
						"$ref": "#/$defs/my_type",
					},
				},
			},
			"definitions": map[string]any{
				"my_type": map[string]any{
					"type": "string",
				},
			},
		}

		processed, err := ProcessDatasourceYAML(yamlData)
		assert.NoError(t, err)

		assert.NotContains(t, processed, "definitions")

		outputSchema := processed["output_schema"].(map[string]any)
		props := outputSchema["properties"].(map[string]any)

		fileRef := props["file"].(map[string]any)
		assert.Equal(t, "object", fileRef["type"])

		customRef := props["custom"].(map[string]any)
		assert.Equal(t, "string", customRef["type"])
	})
}
