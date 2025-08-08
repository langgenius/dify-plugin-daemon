package plugin_entities

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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

// ProcessDataWithRefs processes a data map by resolving schema references
func ProcessDataWithRefs(data map[string]any, schemaKey string) (map[string]any, error) {
	if schema, hasSchema := data[schemaKey]; hasSchema {
		userDefinitions := make(map[string]any)
		if defs, hasDefs := data["definitions"].(map[string]any); hasDefs {
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

		data[schemaKey] = resolvedSchema

		// Remove the definitions section as they have been resolved
		delete(data, "definitions")
	}

	return data, nil
}

func ProcessJSONWithRefs(jsonData []byte, schemaKey string) ([]byte, error) {
	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	processedData, err := ProcessDataWithRefs(data, schemaKey)
	if err != nil {
		return nil, err
	}

	result, err := json.Marshal(processedData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal processed data: %w", err)
	}

	return result, nil
}

// ProcessDatasourceData processes datasource data by resolving output_schema references
func ProcessDatasourceData(data map[string]any) (map[string]any, error) {
	return ProcessDataWithRefs(data, "output_schema")
}

// containsRef recursively checks if a data structure contains any $ref fields
func containsRef(data any) bool {
	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			if key == "$ref" {
				return true
			}
			if containsRef(value) {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if containsRef(item) {
				return true
			}
		}
	}
	return false
}

// ProcessSchema processes a schema by resolving $refs with definitions
// This function is designed to be used directly by custom unmarshaling methods
func ProcessSchema(schema any, definitions map[string]any) (any, error) {
	// If schema doesn't contain any $ref, return it as is
	if !containsRef(schema) {
		return schema, nil
	}

	// Merge builtin definitions with user definitions
	allDefinitions := make(map[string]any)
	for k, v := range BuiltinDefinitions {
		allDefinitions[k] = v
	}
	for k, v := range definitions {
		allDefinitions[k] = v
	}

	return ResolveSchemaRefs(schema, allDefinitions)
}

// ProcessSchemaWithUserDefinitions processes schema and extracts definitions from the same data map
func ProcessSchemaWithUserDefinitions(data map[string]any, schemaKey string) (any, error) {
	schema, hasSchema := data[schemaKey]
	if !hasSchema {
		return nil, nil
	}

	userDefinitions := make(map[string]any)
	if defs, hasDefs := data["definitions"].(map[string]any); hasDefs {
		userDefinitions = defs
	}

	return ProcessSchema(schema, userDefinitions)
}

// ProcessToolData processes tool data by resolving output_schema references
func ProcessToolData(data map[string]any) (map[string]any, error) {
	return ProcessDataWithRefs(data, "output_schema")
}

func ProcessToolJSON(jsonData []byte) ([]byte, error) {
	return ProcessJSONWithRefs(jsonData, "output_schema")
}
