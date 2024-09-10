package plugin_daemon

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func TestToolInvokeJSONSchemaValidator(t *testing.T) {
	response := stream.NewStream[tool_entities.ToolResponseChunk](128)

	bindValidator(response, map[string]any{
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
			},
		},
	})

	response.Write(tool_entities.ToolResponseChunk{
		Type: tool_entities.ToolResponseChunkTypeVariable,
		Message: map[string]any{
			"variable_name":  "name",
			"variable_value": "1",
			"stream":         true,
		},
	})
	response.Close()

	for response.Next() {
		data, err := response.Read()
		if err != nil {
			t.Fatal(err)
		}

		t.Log(data)
	}
}

func TestToolInvokeJSONSchemaValidatorWithInvalidSchema(t *testing.T) {
	response := stream.NewStream[tool_entities.ToolResponseChunk](128)

	bindValidator(response, map[string]any{
		"output_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type": "string",
				},
			},
		},
	})

	response.Write(tool_entities.ToolResponseChunk{
		Type: tool_entities.ToolResponseChunkTypeVariable,
		Message: map[string]any{
			"variable_name":  "name",
			"variable_value": 1,
			"stream":         false,
		},
	})

	response.Close()

	_, err := response.Read()
	if err != nil {
		t.Fatal(err)
	}

	_, err = response.Read()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
