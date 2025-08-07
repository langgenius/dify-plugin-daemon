package builtin_schema

import (
	"errors"
	"fmt"
	"strings"
)

// BuiltinDefinitions contains all the built-in schema definitions that can be referenced
// in plugin YAML files using $refs
var BuiltinDefinitions = map[string]any{
	"file": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"File"},
				"description": "Business type identifier for frontend",
			},
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
	"website_crawl": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"WebsiteCrawl"},
				"description": "Business type identifier for frontend",
			},
			"source_url": map[string]any{
				"type":        "string",
				"description": "The URL of the crawled website",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content of the crawled website",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "The title of the crawled website",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "The description of the crawled website",
			},
		},
		"required": []string{"source_url", "content"},
	},
	"online_document": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"OnlineDocument"},
				"description": "Business type identifier for frontend",
			},
			"workspace_id": map[string]any{
				"type":        "string",
				"description": "The ID of the workspace where the document is stored",
			},
			"page_id": map[string]any{
				"type":        "string",
				"description": "The ID of the page in the document",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content of the online document",
			},
		},
		"required": []string{"content"},
	},
	"general_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"GeneralStructureChunk"},
				"description": "Business type identifier for frontend",
			},
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
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"ParentChildStructureChunk"},
				"description": "Business type identifier for frontend",
			},
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
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"QAStructureChunk"},
				"description": "Business type identifier for frontend",
			},
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

// ResolveSchemaRefs resolves $refs in JSON schema recursively
// It supports references in the format "#/$defs/typename"
func ResolveSchemaRefs(schema any, definitions map[string]any) (any, error) {
	switch v := schema.(type) {
	case map[string]any:
		if ref, hasRef := v["$ref"].(string); hasRef {
			if strings.HasPrefix(ref, "#/$defs/") {
				typeName := strings.TrimPrefix(ref, "#/$defs/")
				if def, exists := definitions[typeName]; exists {
					return ResolveSchemaRefs(def, definitions)
				}
				return nil, fmt.Errorf("reference '%s' not found in definitions", ref)
			}
			return nil, fmt.Errorf("unsupported reference format: %s", ref)
		}

		resolved := make(map[string]any)
		for key, value := range v {
			resolvedValue, err := ResolveSchemaRefs(value, definitions)
			if err != nil {
				return nil, err
			}
			resolved[key] = resolvedValue
		}
		return resolved, nil

	case []any:
		resolved := make([]any, len(v))
		for i, item := range v {
			resolvedItem, err := ResolveSchemaRefs(item, definitions)
			if err != nil {
				return nil, err
			}
			resolved[i] = resolvedItem
		}
		return resolved, nil

	default:
		return schema, nil
	}
}

// ProcessYAMLWithRefs processes a YAML data map by resolving schema references
// in the specified schema key (e.g., "output_schema", "input_schema")
func ProcessYAMLWithRefs(yamlData map[string]any, schemaKey string) (map[string]any, error) {
	if schema, hasSchema := yamlData[schemaKey]; hasSchema {
		userDefinitions := make(map[string]any)
		if defs, hasDefs := yamlData["definitions"].(map[string]any); hasDefs {
			userDefinitions = defs
		}

		// Merge builtin definitions with user definitions
		allDefinitions := make(map[string]any)
		for k, v := range BuiltinDefinitions {
			allDefinitions[k] = v
		}
		for k, v := range userDefinitions {
			allDefinitions[k] = v
		}

		resolvedSchema, err := ResolveSchemaRefs(schema, allDefinitions)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to resolve schema references"))
		}

		yamlData[schemaKey] = resolvedSchema

		// Remove the definitions section as they have been resolved
		delete(yamlData, "definitions")
	}

	return yamlData, nil
}

// ProcessDatasourceYAML processes datasource YAML data by resolving output_schema references
func ProcessDatasourceYAML(yamlData map[string]any) (map[string]any, error) {
	return ProcessYAMLWithRefs(yamlData, "output_schema")
}

// ProcessToolYAML processes tool YAML data by resolving output_schema references
func ProcessToolYAML(yamlData map[string]any) (map[string]any, error) {
	return ProcessYAMLWithRefs(yamlData, "output_schema")
}