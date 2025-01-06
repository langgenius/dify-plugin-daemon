package plugin_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func TestUnmarshalEndpointDeclarationFromYaml(t *testing.T) {
	const data = `settings:
  - name: api_key
    type: secret-input
    required: true
    label:
      en_US: API key
      zh_Hans: API key
      pt_BR: API key
    placeholder:
      en_US: Please input your API key
      zh_Hans: 请输入你的 API key
      pt_BR: Please input your API key
endpoints:
  - endpoints/duck.yaml
  - endpoints/neko.yaml
`

	dec, err := parser.UnmarshalYamlBytes[EndpointProviderDeclaration]([]byte(data))
	if err != nil {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration: %v", err)
	}

	if len(dec.EndpointFiles) != 2 {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration: %v", err)
	}
}

func TestUnmarshalEndpointDeclarationFromYaml2(t *testing.T) {
	const data = `settings:
  - name: api_key
    type: secret-input
    required: true
    label:
      en_US: API key
      zh_Hans: API key
      pt_BR: API key
    placeholder:
      en_US: Please input your API key
      zh_Hans: 请输入你的 API key
      pt_BR: Please input your API key
endpoints:
  - path: "/duck/<app_id>"
    method: "GET"`

	dec, err := parser.UnmarshalYamlBytes[EndpointProviderDeclaration]([]byte(data))
	if err != nil {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration: %v", err)
	}

	if len(dec.Endpoints) != 1 {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration: %v", err)
	}
}

func TestUnmarshalEndpointDeclarationFromJSON(t *testing.T) {
	const data = `{
		"settings": [
			{
				"name": "api_key",
				"type": "secret-input",
				"required": true,
				"label": {
					"en_US": "API key",
					"zh_Hans": "API key",
					"pt_BR": "API key"
				},
				"placeholder": {
					"en_US": "Please input your API key",
					"zh_Hans": "请输入你的 API key",
					"pt_BR": "Please input your API key"
				}
			}
		],
		"endpoints": [
			"endpoints/duck.yaml",
			"endpoints/neko.yaml"
		]
	}`

	dec, err := parser.UnmarshalJsonBytes[EndpointProviderDeclaration]([]byte(data))
	if err != nil {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration from JSON: %v", err)
	}

	if len(dec.EndpointFiles) != 2 {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration from JSON: expected 1 endpoint, got %d", len(dec.Endpoints))
	}

	if len(dec.Settings) != 1 {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration from JSON: expected 1 setting, got %d", len(dec.Settings))
	}
}

func TestUnmarshalEndpointDeclarationFromJSON2(t *testing.T) {
	const data = `{
		"settings": [
			{
				"name": "api_key",
				"type": "secret-input",
				"required": true,
				"label": {
					"en_US": "API key",
					"zh_Hans": "API key",
					"pt_BR": "API key"
				},
				"placeholder": {
					"en_US": "Please input your API key",
					"zh_Hans": "请输入你的 API key",
					"pt_BR": "Please input your API key"
				}
			}
		],
		"endpoints": [
			{
				"path": "/duck/<app_id>",
				"method": "GET"
			}
		]
	}`

	dec, err := parser.UnmarshalJsonBytes[EndpointProviderDeclaration]([]byte(data))
	if err != nil {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration from JSON: %v", err)
	}

	if len(dec.Endpoints) != 1 {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration from JSON: expected 1 endpoint, got %d", len(dec.Endpoints))
	}

	if len(dec.Settings) != 1 {
		t.Fatalf("Failed to unmarshal EndpointProviderDeclaration from JSON: expected 1 setting, got %d", len(dec.Settings))
	}
}
