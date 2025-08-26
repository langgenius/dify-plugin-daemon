package plugin_entities

import (
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

// TestDatasourceDeclarationBasic tests basic datasource declaration marshaling/unmarshaling
func TestDatasourceDeclarationBasic(t *testing.T) {
	jsonData := `{
		"identity": {
			"author": "test_author",
			"name": "test_datasource",
			"label": {
				"en_US": "Test Datasource",
				"zh_Hans": "测试数据源"
			},
			"icon": "icon.svg"
		},
		"parameters": [
			{
				"name": "api_key",
				"type": "secret-input",
				"label": {
					"en_US": "API Key",
					"zh_Hans": "API密钥"
				},
				"description": {
					"en_US": "Your API key",
					"zh_Hans": "您的API密钥"
				},
				"required": true
			}
		],
		"description": {
			"en_US": "A test datasource",
			"zh_Hans": "测试数据源"
		}
	}`

	datasourceDecl, err := parser.UnmarshalJsonBytes[DatasourceDeclaration]([]byte(jsonData))
	if err != nil {
		t.Errorf("Failed to unmarshal basic datasource declaration: %v", err)
	}

	if datasourceDecl.Identity.Author != "test_author" {
		t.Errorf("Expected author 'test_author', got '%s'", datasourceDecl.Identity.Author)
	}

	if len(datasourceDecl.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(datasourceDecl.Parameters))
	}
}


// TestDatasourceProviderDeclaration tests the full provider declaration
func TestDatasourceProviderDeclaration(t *testing.T) {
	yamlData := `
identity:
  author: test_author
  name: test_provider
  description:
    en_US: Test Provider
    zh_Hans: 测试提供者
  icon: icon.svg
  label:
    en_US: Test Provider
    zh_Hans: 测试提供者
  tags:
    - search
    - analysis
credentials_schema:
  - name: api_key
    type: secret-input
    required: true
    label:
      en_US: API Key
    help:
      en_US: Your API key
provider_type: online_document
datasources:
  - identity:
      author: test_author
      name: datasource1
      label:
        en_US: Datasource 1
    parameters: []
    description:
      en_US: First datasource
    output_schema:
      type: object
      properties:
        result:
          type: object
`

	providerDecl, err := parser.UnmarshalYamlBytes[DatasourceProviderDeclaration]([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal provider declaration: %v", err)
	}

	if providerDecl.Identity.Author != "test_author" {
		t.Errorf("Expected author 'test_author', got '%s'", providerDecl.Identity.Author)
	}

	if providerDecl.ProviderType != DatasourceTypeOnlineDocument {
		t.Errorf("Expected provider type 'online_document', got '%s'", providerDecl.ProviderType)
	}

	if len(providerDecl.Datasources) != 1 {
		t.Fatalf("Expected 1 datasource, got %d", len(providerDecl.Datasources))
	}

}


// TestDatasourceParameterValidation tests parameter validation
func TestDatasourceParameterValidation(t *testing.T) {
	// Test valid parameter types
	validTypes := []string{"string", "number", "boolean", "select", "secret-input"}
	for _, paramType := range validTypes {
		jsonData := `{
			"identity": {
				"author": "test",
				"name": "test",
				"label": {"en_US": "Test"}
			},
			"parameters": [{
				"name": "param",
				"type": "` + paramType + `",
				"label": {"en_US": "Param"},
				"description": {"en_US": "Description"},
				"required": false
			}],
			"description": {"en_US": "Test"}
		}`

		_, err := parser.UnmarshalJsonBytes[DatasourceDeclaration]([]byte(jsonData))
		if err != nil {
			t.Errorf("Failed to unmarshal datasource with %s parameter type: %v", paramType, err)
		}
	}

	// Test invalid parameter type
	invalidJsonData := `{
		"identity": {
			"author": "test",
			"name": "test",
			"label": {"en_US": "Test"}
		},
		"parameters": [{
			"name": "param",
			"type": "invalid_type",
			"label": {"en_US": "Param"},
			"description": {"en_US": "Description"},
			"required": false
		}],
		"description": {"en_US": "Test"}
	}`

	_, err := parser.UnmarshalJsonBytes[DatasourceDeclaration]([]byte(invalidJsonData))
	if err == nil {
		t.Error("Should fail with invalid parameter type")
	}
}

// TestDatasourceProviderTypeValidation tests provider type validation
func TestDatasourceProviderTypeValidation(t *testing.T) {
	// Test valid provider types
	validTypes := []string{"website_crawl", "online_document", "online_drive"}
	for _, providerType := range validTypes {
		yamlData := `
identity:
  author: test
  name: test
  description:
    en_US: Test
  icon: icon.svg
  label:
    en_US: Test
provider_type: ` + providerType + `
datasources: []
`

		_, err := parser.UnmarshalYamlBytes[DatasourceProviderDeclaration]([]byte(yamlData))
		if err != nil {
			t.Errorf("Failed to unmarshal provider with %s type: %v", providerType, err)
		}
	}

	// Test invalid provider type
	invalidYamlData := `
identity:
  author: test
  name: test
  description:
    en_US: Test
  icon: icon.svg
  label:
    en_US: Test
provider_type: invalid_type
datasources: []
`

	decl, err := parser.UnmarshalYamlBytes[DatasourceProviderDeclaration]([]byte(invalidYamlData), *validators.GlobalEntitiesValidator)
	if err == nil {
		// If unmarshaling succeeded, manually validate
		err = validators.GlobalEntitiesValidator.Struct(decl)
		if err == nil {
			t.Error("Should fail with invalid provider type")
		} else if !strings.Contains(err.Error(), "datasource_provider_type") && !strings.Contains(err.Error(), "provider_type") {
			t.Errorf("Error should mention invalid provider type, got: %v", err)
		}
	}
}
