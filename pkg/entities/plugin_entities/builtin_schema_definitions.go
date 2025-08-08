package plugin_entities

// BuiltinDefinitions contains all the built-in schema definitions that can be referenced
// in plugin YAML files using $refs
var BuiltinDefinitions = map[string]any{
	"file": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "file name",
			},
			"size": map[string]any{
				"type":        "number",
				"description": "file size",
			},
			"file_type": map[string]any{
				"type":        "string",
				"description": "file type",
			},
			"extension": map[string]any{
				"type":        "string",
				"description": "file extension",
			},
			"mime_type": map[string]any{
				"type":        "string",
				"description": "file mime type",
			},
			"transfer_method": map[string]any{
				"type":        "string",
				"description": "file transfer method",
			},
			"url": map[string]any{
				"type":        "string",
				"description": "file url",
			},
			"related_id": map[string]any{
				"type":        "string",
				"description": "file related id",
			},
		},
		"required": []string{"name"},
	},
	"general_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"general_chunks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "List of general content chunks",
			},
		},
		"required": []string{"general_chunks"},
	},
	"parent_child_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"parent_mode": map[string]any{
				"type":        "string",
				"description": "The mode of parent-child relationship",
			},
			"parent_child_chunks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"parent_content": map[string]any{
							"type":        "string",
							"description": "The parent content",
						},
						"child_contents": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
							"description": "List of child contents",
						},
					},
					"required": []string{"parent_content", "child_contents"},
				},
				"description": "List of parent-child chunk pairs",
			},
		},
		"required": []string{"parent_mode", "parent_child_chunks"},
	},
	"qa_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"qa_chunks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]any{
							"type":        "string",
							"description": "The question",
						},
						"answer": map[string]any{
							"type":        "string",
							"description": "The answer",
						},
					},
					"required": []string{"question", "answer"},
				},
				"description": "List of question-answer pairs",
			},
		},
		"required": []string{"qa_chunks"},
	},
}