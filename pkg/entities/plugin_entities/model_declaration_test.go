package plugin_entities

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"gopkg.in/yaml.v3"
)

func parse_yaml_to_json(data []byte) ([]byte, error) {
	var obj interface{}
	err := yaml.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func TestModelDeclarationNormalizeModelProperties(t *testing.T) {
	model := ModelDeclaration{
		ModelProperties: map[string]any{
			"nested": map[any]any{
				"int_key": map[any]any{
					1: "one",
				},
				"slice": []any{
					map[any]any{
						2: "two",
					},
				},
			},
		},
	}

	model.normalizeModelProperties()

	expected := map[string]any{
		"nested": map[string]any{
			"int_key": map[string]any{
				"1": "one",
			},
			"slice": []any{
				map[string]any{
					"2": "two",
				},
			},
		},
	}
	if !reflect.DeepEqual(model.ModelProperties, expected) {
		t.Fatalf("unexpected normalized model properties: %#v", model.ModelProperties)
	}
}

func TestFullFunctionModelProvider_Validate(t *testing.T) {
	const (
		model_provider_template = `
    provider: openai
    label:
      en_US: OpenAI
    description:
      en_US: Models provided by OpenAI, such as GPT-3.5-Turbo and GPT-4.
      zh_Hans: OpenAI 提供的模型，例如 GPT-3.5-Turbo 和 GPT-4。
    icon_small:
      en_US: icon_s_en.svg
    icon_large:
      en_US: icon_l_en.svg
    background: "#E5E7EB"
    help:
      title:
        en_US: Get your API Key from OpenAI
        zh_Hans: 从 OpenAI 获取 API Key
      url:
        en_US: https://platform.openai.com/account/api-keys
    supported_model_types:
      - llm
      - text-embedding
      - speech2text
      - moderation
      - tts
    configurate_methods:
      - predefined-model
      - customizable-model
    model_credential_schema:
      model:
        label:
          en_US: Model Name
          zh_Hans: 模型名称
        placeholder:
          en_US: Enter your model name
          zh_Hans: 输入模型名称
      credential_form_schemas:
        - variable: openai_api_key
          label:
            en_US: API Key
          type: secret-input
          required: true
          placeholder:
            zh_Hans: 在此输入您的 API Key
            en_US: Enter your API Key
        - variable: openai_organization
          label:
            zh_Hans: 组织 ID
            en_US: Organization
          type: text-input
          required: false
          placeholder:
            zh_Hans: 在此输入您的组织 ID
            en_US: Enter your Organization ID
        - variable: openai_api_base
          label:
            zh_Hans: API Base
            en_US: API Base
          type: text-input
          required: false
          placeholder:
            zh_Hans: 在此输入您的 API Base
            en_US: Enter your API Base
    provider_credential_schema:
      credential_form_schemas:
        - variable: openai_api_key
          label:
            en_US: API Key
          type: secret-input
          required: true
          placeholder:
            zh_Hans: 在此输入您的 API Key
            en_US: Enter your API Key
        - variable: openai_organization
          label:
            zh_Hans: 组织 ID
            en_US: Organization
          type: text-input
          required: false
          placeholder:
            zh_Hans: 在此输入您的组织 ID
            en_US: Enter your Organization ID
        - variable: openai_api_base
          label:
            zh_Hans: API Base
            en_US: API Base
          type: text-input
          required: false
          placeholder:
            zh_Hans: 在此输入您的 API Base, 如：https://api.openai.com
            en_US: Enter your API Base, e.g. https://api.openai.com
        `
	)
	jsonData, err := parse_yaml_to_json([]byte(model_provider_template))
	if err != nil {
		t.Error(err)
	}

	_, err = parser.UnmarshalYamlBytes[ModelProviderDeclaration](jsonData)
	if err != nil {
		t.Errorf("UnmarshalModelProviderConfiguration() error = %v", err)
	}
}

func TestModelParameterRule_UseTemplateYAML(t *testing.T) {
	const (
		model_parameter_rule_template = `
name: temperature
use_template: temperature
`
	)

	yamlData := []byte(model_parameter_rule_template)

	model, err := parser.UnmarshalYamlBytes[ModelParameterRule](yamlData)
	if err != nil {
		t.Errorf("UnmarshalModelParameterRule() error = %v", err)
		return
	}

	if model.Type == nil {
		t.Errorf("UnmarshalModelParameterRule() error = %v", err)
		return
	}

	if *model.Type != PARAMETER_TYPE_FLOAT {
		t.Errorf("UnmarshalModelParameterRule() error = %v", err)
	}

	if model.Min == nil || model.Max == nil || model.Precision == nil {
		t.Errorf("Missing default value")
	}
}

func TestModelParameterRule_UseTemplateJSON(t *testing.T) {
	const (
		model_parameter_rule_template = `{"name": "temperature", "use_template": "temperature"}`
	)

	jsonData := []byte(model_parameter_rule_template)

	model, err := parser.UnmarshalJsonBytes[ModelParameterRule](jsonData)
	if err != nil {
		t.Errorf("UnmarshalModelParameterRule() error = %v", err)
	}

	if model.Type == nil {
		t.Errorf("UnmarshalModelParameterRule() error = %v", err)
	}

	if *model.Type != PARAMETER_TYPE_FLOAT {
		t.Errorf("UnmarshalModelParameterRule() error = %v", err)
	}

	if model.Min == nil || model.Max == nil || model.Precision == nil {
		t.Errorf("Missing default value")
	}
}

func TestModelProviderCredentialFormSchema_HelpSurvivesRoundTrip(t *testing.T) {
	const model_provider_template = `
    provider: openai
    label:
      en_US: OpenAI
    provider_credential_schema:
      credential_form_schemas:
        - variable: openai_api_base
          label:
            en_US: API Base
          type: text-input
          required: false
          help:
            en_US: Custom API base URL, e.g. a proxy or gateway
            zh_Hans: 自定义 API 基础地址，例如代理或网关
          url: https://platform.openai.com/docs/api-reference
          placeholder:
            en_US: Enter your API Base
        `

	jsonData, err := parse_yaml_to_json([]byte(model_provider_template))
	if err != nil {
		t.Fatal(err)
	}

	declaration, err := parser.UnmarshalYamlBytes[ModelProviderDeclaration](jsonData)
	if err != nil {
		t.Fatalf("UnmarshalYamlBytes() error = %v", err)
	}

	if declaration.ProviderCredentialSchema == nil {
		t.Fatal("provider_credential_schema was not parsed")
	}
	schemas := declaration.ProviderCredentialSchema.CredentialFormSchemas
	if len(schemas) != 1 {
		t.Fatalf("expected 1 credential form schema, got %d", len(schemas))
	}
	field := schemas[0]

	if field.Help == nil {
		t.Fatal("Help was dropped during unmarshal; expected the manifest help to be retained")
	}
	if field.Help.EnUS != "Custom API base URL, e.g. a proxy or gateway" {
		t.Errorf("Help.EnUS = %q, want the manifest help text", field.Help.EnUS)
	}
	if field.URL == nil || *field.URL != "https://platform.openai.com/docs/api-reference" {
		t.Errorf("URL = %v, want the manifest url", field.URL)
	}

	roundTripped, err := json.Marshal(declaration)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(roundTripped, []byte(`"help":`)) {
		t.Error("marshaled JSON does not contain a help key")
	}
	if !bytes.Contains(roundTripped, []byte("Custom API base URL, e.g. a proxy or gateway")) {
		t.Error("marshaled JSON does not retain the help text")
	}
}
