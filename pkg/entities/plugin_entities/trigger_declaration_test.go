package plugin_entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConvertYAMLNodeToProviderConfigList(t *testing.T) {
	tests := []struct {
		name     string
		yamlStr  string
		expected []ProviderConfig
		wantErr  bool
	}{
		{
			name: "array format",
			yamlStr: `
- name: api_key
  type: string
  required: true
- name: endpoint
  type: string
  required: false`,
			expected: []ProviderConfig{
				{Name: "api_key", Type: "string", Required: true},
				{Name: "endpoint", Type: "string", Required: false},
			},
			wantErr: false,
		},
		{
			name: "dict format",
			yamlStr: `
api_key:
  type: string
  required: true
endpoint:
  type: string
  required: false`,
			expected: []ProviderConfig{
				{Name: "api_key", Type: "string", Required: true},
				{Name: "endpoint", Type: "string", Required: false},
			},
			wantErr: false,
		},
		{
			name:     "empty array",
			yamlStr:  `[]`,
			expected: []ProviderConfig{},
			wantErr:  false,
		},
		{
			name:     "empty dict",
			yamlStr:  `{}`,
			expected: []ProviderConfig{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yamlStr), &node)
			assert.NoError(t, err)

			// The unmarshal creates a Document node, we need its content
			contentNode := node.Content[0]

			result, err := convertYAMLNodeToProviderConfigList(contentNode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTriggerProviderDeclaration_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yamlStr string
		check   func(t *testing.T, declaration *TriggerProviderDeclaration)
		wantErr bool
	}{
		{
			name: "subscription_schema with dict format",
			yamlStr: `
identity:
  name: test_provider
  author: test_author
subscription_schema:
  api_key:
    type: string
    required: true
  secret:
    type: password
    required: true`,
			check: func(t *testing.T, d *TriggerProviderDeclaration) {
				assert.Len(t, d.SubscriptionSchema, 2)
				assert.Equal(t, "api_key", d.SubscriptionSchema[0].Name)
				assert.Equal(t, ConfigType("string"), d.SubscriptionSchema[0].Type)
				assert.True(t, d.SubscriptionSchema[0].Required)
				assert.Equal(t, "secret", d.SubscriptionSchema[1].Name)
				assert.Equal(t, ConfigType("password"), d.SubscriptionSchema[1].Type)
			},
			wantErr: false,
		},
		{
			name: "subscription_constructor.credentials_schema with dict format",
			yamlStr: `
identity:
  name: test_provider
  author: test_author
subscription_schema:
  - name: webhook_url
    type: string
    required: true
subscription_constructor:
  credentials_schema:
    client_id:
      type: string
      required: true
    client_secret:
      type: password
      required: true`,
			check: func(t *testing.T, d *TriggerProviderDeclaration) {
				assert.NotNil(t, d.SubscriptionConstructor)
				assert.Len(t, d.SubscriptionConstructor.CredentialsSchema, 2)
				assert.Equal(t, "client_id", d.SubscriptionConstructor.CredentialsSchema[0].Name)
				assert.Equal(t, ConfigType("string"), d.SubscriptionConstructor.CredentialsSchema[0].Type)
				assert.Equal(t, "client_secret", d.SubscriptionConstructor.CredentialsSchema[1].Name)
				assert.Equal(t, ConfigType("password"), d.SubscriptionConstructor.CredentialsSchema[1].Type)
			},
			wantErr: false,
		},
		{
			name: "mixed formats - dict subscription_schema and array credentials_schema",
			yamlStr: `
identity:
  name: test_provider
  author: test_author
subscription_schema:
  webhook_url:
    type: string
    required: true
subscription_constructor:
  credentials_schema:
    - name: api_key
      type: string
      required: true`,
			check: func(t *testing.T, d *TriggerProviderDeclaration) {
				assert.Len(t, d.SubscriptionSchema, 1)
				assert.Equal(t, "webhook_url", d.SubscriptionSchema[0].Name)
				assert.NotNil(t, d.SubscriptionConstructor)
				assert.Len(t, d.SubscriptionConstructor.CredentialsSchema, 1)
				assert.Equal(t, "api_key", d.SubscriptionConstructor.CredentialsSchema[0].Name)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var declaration TriggerProviderDeclaration
			err := yaml.Unmarshal([]byte(tt.yamlStr), &declaration)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.check(t, &declaration)
			}
		})
	}
}
