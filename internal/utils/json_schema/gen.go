package jsonschema

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"github.com/xeipuuv/gojsonschema"
)

// generate a validate json from its json schema
func GenerateValidateJson(schema map[string]any) (map[string]any, error) {
	// We need to use the schema object to generate a valid JSON
	// Since gojsonschema doesn't provide a direct way to generate valid JSON from a schema,
	// we'll need to implement our own logic based on the schema structure

	result := map[string]any{}

	// Get the schema type
	schemaLoader := gojsonschema.NewGoLoader(schema)
	schemaDoc, err := schemaLoader.LoadJSON()
	if err != nil {
		return nil, err
	}

	// Process the schema document to generate valid JSON
	if schemaType, ok := schemaDoc.(map[string]interface{})["type"].(string); ok && schemaType == "object" {
		// Get properties
		if properties, ok := schemaDoc.(map[string]interface{})["properties"].(map[string]interface{}); ok {
			for propName, propSchema := range properties {
				propSchemaMap, ok := propSchema.(map[string]interface{})
				if !ok {
					continue
				}

				propType, _ := propSchemaMap["type"].(string)
				switch propType {
				case "string":
					// Check for enum values
					if enum, ok := propSchemaMap["enum"].([]interface{}); ok && len(enum) > 0 {
						// Use first enum value
						result[propName] = enum[0]
					} else {
						// Generate random string
						result[propName] = strings.RandomString(10)
					}
				case "number", "integer":
					// Check for enum values
					if enum, ok := propSchemaMap["enum"].([]interface{}); ok && len(enum) > 0 {
						// Use first enum value
						result[propName] = enum[0]
					} else {
						// Generate random number
						min := 0.0
						max := 100.0

						if minVal, ok := propSchemaMap["minimum"].(float64); ok {
							min = minVal
						}
						if maxVal, ok := propSchemaMap["maximum"].(float64); ok {
							max = maxVal
						}

						// Simple random number between min and max
						result[propName] = min + (max-min)/2
					}
				case "boolean":
					// Default to true
					result[propName] = true
				case "array":
					// Create an empty array
					arr := []interface{}{}

					// If items are defined, add a sample item
					if items, ok := propSchemaMap["items"].(map[string]interface{}); ok {
						itemType, _ := items["type"].(string)
						switch itemType {
						case "string":
							arr = append(arr, "sample_item")
						case "number", "integer":
							arr = append(arr, 42)
						case "boolean":
							arr = append(arr, true)
						}
					}

					result[propName] = arr
				case "object":
					// Create a nested object
					nestedObj := map[string]interface{}{}

					if nestedProps, ok := propSchemaMap["properties"].(map[string]interface{}); ok {
						for nestedPropName, nestedPropSchema := range nestedProps {
							if nestedPropMap, ok := nestedPropSchema.(map[string]interface{}); ok {
								nestedType, _ := nestedPropMap["type"].(string)
								switch nestedType {
								case "string":
									nestedObj[nestedPropName] = "nested_" + nestedPropName
								case "number", "integer":
									nestedObj[nestedPropName] = 42
								case "boolean":
									nestedObj[nestedPropName] = true
								}
							}
						}
					}

					result[propName] = nestedObj
				}
			}
		}
	}

	return result, nil
}
