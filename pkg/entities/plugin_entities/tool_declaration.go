package plugin_entities

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"gopkg.in/yaml.v3"
)

type ToolIdentity struct {
	Author string     `json:"author" yaml:"author" validate:"required"`
	Name   string     `json:"name" yaml:"name" validate:"required,tool_identity_name"`
	Label  I18nObject `json:"label" yaml:"label" validate:"required"`
}

var toolIdentityNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isToolIdentityName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return toolIdentityNameRegex.MatchString(value)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("tool_identity_name", isToolIdentityName)
}

type ToolParameterType string

const (
	TOOL_PARAMETER_TYPE_STRING         ToolParameterType = STRING
	TOOL_PARAMETER_TYPE_NUMBER         ToolParameterType = NUMBER
	TOOL_PARAMETER_TYPE_BOOLEAN        ToolParameterType = BOOLEAN
	TOOL_PARAMETER_TYPE_SELECT         ToolParameterType = SELECT
	TOOL_PARAMETER_TYPE_SECRET_INPUT   ToolParameterType = SECRET_INPUT
	TOOL_PARAMETER_TYPE_FILE           ToolParameterType = FILE
	TOOL_PARAMETER_TYPE_FILES          ToolParameterType = FILES
	TOOL_PARAMETER_TYPE_APP_SELECTOR   ToolParameterType = APP_SELECTOR
	TOOL_PARAMETER_TYPE_MODEL_SELECTOR ToolParameterType = MODEL_SELECTOR
	// TOOL_PARAMETER_TYPE_TOOL_SELECTOR  ToolParameterType = TOOL_SELECTOR
	TOOL_PARAMETER_TYPE_ANY            ToolParameterType = ANY
	TOOL_PARAMETER_TYPE_DYNAMIC_SELECT ToolParameterType = DYNAMIC_SELECT
	TOOL_PARAMETER_ARRAY               ToolParameterType = ARRAY
	TOOL_PARAMETER_OBJECT              ToolParameterType = OBJECT
)

func isToolParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(TOOL_PARAMETER_TYPE_STRING),
		string(TOOL_PARAMETER_TYPE_NUMBER),
		string(TOOL_PARAMETER_TYPE_BOOLEAN),
		string(TOOL_PARAMETER_TYPE_SELECT),
		string(TOOL_PARAMETER_TYPE_SECRET_INPUT),
		string(TOOL_PARAMETER_TYPE_FILE),
		string(TOOL_PARAMETER_TYPE_FILES),
		// string(TOOL_PARAMETER_TYPE_TOOL_SELECTOR),
		string(TOOL_PARAMETER_TYPE_APP_SELECTOR),
		string(TOOL_PARAMETER_TYPE_MODEL_SELECTOR),
		string(TOOL_PARAMETER_TYPE_ANY),
		string(TOOL_PARAMETER_TYPE_DYNAMIC_SELECT),
		string(TOOL_PARAMETER_ARRAY),
		string(TOOL_PARAMETER_OBJECT):
		return true
	}
	return false
}

type ToolParameterForm string

const (
	TOOL_PARAMETER_FORM_SCHEMA ToolParameterForm = "schema"
	TOOL_PARAMETER_FORM_FORM   ToolParameterForm = "form"
	TOOL_PARAMETER_FORM_LLM    ToolParameterForm = "llm"
)

func isToolParameterForm(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(TOOL_PARAMETER_FORM_SCHEMA),
		string(TOOL_PARAMETER_FORM_FORM),
		string(TOOL_PARAMETER_FORM_LLM):
		return true
	}
	return false
}

type ParameterAutoGenerateType string

const (
	PARAMETER_AUTO_GENERATE_TYPE_PROMPT_INSTRUCTION ParameterAutoGenerateType = "prompt_instruction"
)

func isParameterAutoGenerateType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PARAMETER_AUTO_GENERATE_TYPE_PROMPT_INSTRUCTION):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("parameter_auto_generate_type", isParameterAutoGenerateType)
}

type ParameterAutoGenerate struct {
	Type ParameterAutoGenerateType `json:"type" yaml:"type" validate:"required,parameter_auto_generate_type"`
}

type ParameterTemplate struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type ToolParameter struct {
	Name             string                 `json:"name" yaml:"name" validate:"required,gt=0,lt=1024"`
	Label            I18nObject             `json:"label" yaml:"label" validate:"required"`
	HumanDescription I18nObject             `json:"human_description" yaml:"human_description" validate:"required"`
	Type             ToolParameterType      `json:"type" yaml:"type" validate:"required,tool_parameter_type"`
	Scope            *string                `json:"scope" yaml:"scope" validate:"omitempty,max=1024,is_scope"`
	Form             ToolParameterForm      `json:"form" yaml:"form" validate:"required,tool_parameter_form"`
	LLMDescription   string                 `json:"llm_description" yaml:"llm_description" validate:"omitempty"`
	Required         bool                   `json:"required" yaml:"required"`
	AutoGenerate     *ParameterAutoGenerate `json:"auto_generate" yaml:"auto_generate" validate:"omitempty"`
	Template         *ParameterTemplate     `json:"template" yaml:"template" validate:"omitempty"`
	Default          any                    `json:"default" yaml:"default" validate:"omitempty,is_basic_type"`
	Min              *float64               `json:"min" yaml:"min" validate:"omitempty"`
	Max              *float64               `json:"max" yaml:"max" validate:"omitempty"`
	Precision        *int                   `json:"precision" yaml:"precision" validate:"omitempty"`
	Options          []ParameterOption      `json:"options" yaml:"options" validate:"omitempty,dive"`
}

type ToolDescription struct {
	Human I18nObject `json:"human" validate:"required"`
	LLM   string     `json:"llm" validate:"required"`
}

type ToolOutputSchema map[string]any

// UnmarshalYAML handles YAML unmarshaling
func (t *ToolOutputSchema) UnmarshalYAML(value *yaml.Node) error {
	var rawData map[string]any
	if err := value.Decode(&rawData); err != nil {
		return err
	}
	*t = ToolOutputSchema(rawData)
	return nil
}

// UnmarshalJSON handles JSON unmarshaling
func (t *ToolOutputSchema) UnmarshalJSON(data []byte) error {
	var temp map[string]any
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	*t = ToolOutputSchema(temp)
	return nil
}

type ToolDeclaration struct {
	Identity             ToolIdentity     `json:"identity" yaml:"identity" validate:"required"`
	Description          ToolDescription  `json:"description" yaml:"description" validate:"required"`
	Parameters           []ToolParameter  `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	OutputSchema         ToolOutputSchema `json:"output_schema,omitempty" yaml:"output_schema,omitempty"`
	HasRuntimeParameters bool             `json:"has_runtime_parameters" yaml:"has_runtime_parameters"`
}

type ToolProviderIdentity struct {
	Author      string                        `json:"author" validate:"required"`
	Name        string                        `json:"name" validate:"required,tool_provider_identity_name"`
	Description I18nObject                    `json:"description"`
	Icon        string                        `json:"icon" validate:"required"`
	IconDark    string                        `json:"icon_dark" validate:"omitempty"`
	Label       I18nObject                    `json:"label" validate:"required"`
	Tags        []manifest_entities.PluginTag `json:"tags" validate:"omitempty,dive,plugin_tag"`
}

var toolProviderIdentityNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isToolProviderIdentityName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return toolProviderIdentityNameRegex.MatchString(value)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("tool_provider_identity_name", isToolProviderIdentityName)
}

type ToolProviderDeclaration struct {
	Identity          ToolProviderIdentity `json:"identity" yaml:"identity" validate:"required"`
	CredentialsSchema []ProviderConfig     `json:"credentials_schema" yaml:"credentials_schema" validate:"omitempty,dive"`
	OAuthSchema       *OAuthSchema         `json:"oauth_schema" yaml:"oauth_schema" validate:"omitempty"`
	Tools             []ToolDeclaration    `json:"tools" yaml:"tools" validate:"required,dive"`
	ToolFiles         []string             `json:"-" yaml:"-"`
}

func (t *ToolProviderDeclaration) MarshalJSON() ([]byte, error) {
	type alias ToolProviderDeclaration
	p := alias(*t)
	if p.CredentialsSchema == nil {
		p.CredentialsSchema = []ProviderConfig{}
	}
	if p.Tools == nil {
		p.Tools = []ToolDeclaration{}
	}
	return json.Marshal(p)
}

func (t *ToolProviderDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type alias struct {
		Identity               ToolProviderIdentity `yaml:"identity"`
		CredentialsSchema      yaml.Node            `yaml:"credentials_schema"`
		CredentialsForProvider yaml.Node            `yaml:"credentials_for_provider"`
		Tools                  yaml.Node            `yaml:"tools"`
		OAuthSchema            *OAuthSchema         `yaml:"oauth_schema"`
	}

	var temp alias

	err := value.Decode(&temp)
	if err != nil {
		return err
	}

	// apply credentials_for_provider to credentials_schema if not exists
	if (temp.CredentialsSchema.Kind == yaml.ScalarNode && temp.CredentialsSchema.Value == "") ||
		len(temp.CredentialsSchema.Content) == 0 {
		temp.CredentialsSchema = temp.CredentialsForProvider
	}

	// apply identity
	t.Identity = temp.Identity

	// apply oauth_schema
	t.OAuthSchema = temp.OAuthSchema

	// check if credentials_schema is a map
	if temp.CredentialsSchema.Kind != yaml.MappingNode {
		// not a map, convert it into array
		credentialsSchema := make([]ProviderConfig, 0)
		if err := temp.CredentialsSchema.Decode(&credentialsSchema); err != nil {
			return err
		}
		t.CredentialsSchema = credentialsSchema
	} else if temp.CredentialsSchema.Kind == yaml.MappingNode {
		credentialsSchema := make([]ProviderConfig, 0, len(temp.CredentialsSchema.Content)/2)
		currentKey := ""
		currentValue := &ProviderConfig{}
		for _, item := range temp.CredentialsSchema.Content {
			if item.Kind == yaml.ScalarNode {
				currentKey = item.Value
			} else if item.Kind == yaml.MappingNode {
				currentValue = &ProviderConfig{}
				if err := item.Decode(currentValue); err != nil {
					return err
				}
				currentValue.Name = currentKey
				credentialsSchema = append(credentialsSchema, *currentValue)
			}
		}
		t.CredentialsSchema = credentialsSchema
	}

	if t.ToolFiles == nil {
		t.ToolFiles = []string{}
	}

	// unmarshal tools
	if temp.Tools.Kind == yaml.SequenceNode {
		for _, item := range temp.Tools.Content {
			if item.Kind == yaml.ScalarNode {
				t.ToolFiles = append(t.ToolFiles, item.Value)
			} else if item.Kind == yaml.MappingNode {
				tool := ToolDeclaration{}
				if err := item.Decode(&tool); err != nil {
					return err
				}
				t.Tools = append(t.Tools, tool)
			}
		}
	}

	if t.CredentialsSchema == nil {
		t.CredentialsSchema = []ProviderConfig{}
	}

	if t.Tools == nil {
		t.Tools = []ToolDeclaration{}
	}

	if t.Identity.Tags == nil {
		t.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}

func (t *ToolProviderDeclaration) UnmarshalJSON(data []byte) error {
	type alias ToolProviderDeclaration

	var temp struct {
		alias
		CredentialsSchema      json.RawMessage   `json:"credentials_schema"`
		CredentialsForProvider json.RawMessage   `json:"credentials_for_provider"`
		Tools                  []json.RawMessage `json:"tools"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.CredentialsSchema) == 0 {
		temp.CredentialsSchema = temp.CredentialsForProvider
	}

	*t = ToolProviderDeclaration(temp.alias)

	// Determine the type of CredentialsSchema
	var raw_message map[string]json.RawMessage
	if err := json.Unmarshal(temp.CredentialsSchema, &raw_message); err == nil {
		// It's an object
		credentialsSchemaObject := make(map[string]ProviderConfig)
		if err := json.Unmarshal(temp.CredentialsSchema, &credentialsSchemaObject); err != nil {
			return fmt.Errorf("failed to unmarshal credentials_schema as object: %v", err)
		}
		for _, value := range credentialsSchemaObject {
			t.CredentialsSchema = append(t.CredentialsSchema, value)
		}
	} else {
		// It's likely an array
		var credentials_schema_array []ProviderConfig
		if err := json.Unmarshal(temp.CredentialsSchema, &credentials_schema_array); err != nil {
			return fmt.Errorf("failed to unmarshal credentials_schema as array: %v", err)
		}
		t.CredentialsSchema = credentials_schema_array
	}

	if t.ToolFiles == nil {
		t.ToolFiles = []string{}
	}

	// unmarshal tools
	for _, item := range temp.Tools {
		tool := ToolDeclaration{}
		if err := json.Unmarshal(item, &tool); err != nil {
			// try to unmarshal it as a string directly
			t.ToolFiles = append(t.ToolFiles, string(item))
		} else {
			t.Tools = append(t.Tools, tool)
		}
	}

	if t.CredentialsSchema == nil {
		t.CredentialsSchema = []ProviderConfig{}
	}

	if t.Tools == nil {
		t.Tools = []ToolDeclaration{}
	}

	if t.Identity.Tags == nil {
		t.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}

func init() {
	// init validator
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")
	// register translations for default validators
	en_translations.RegisterDefaultTranslations(validators.GlobalEntitiesValidator, translator)

	validators.GlobalEntitiesValidator.RegisterValidation("tool_parameter_type", isToolParameterType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"tool_parameter_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("tool_parameter_type", "{0} is not a valid tool parameter type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("tool_parameter_type", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("tool_parameter_form", isToolParameterForm)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"tool_parameter_form",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("tool_parameter_form", "{0} is not a valid tool parameter form", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("tool_parameter_form", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("is_basic_type", isBasicType)
}

func UnmarshalToolProviderDeclaration(data []byte) (*ToolProviderDeclaration, error) {
	obj, err := parser.UnmarshalJsonBytes[ToolProviderDeclaration](data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool provider configuration: %w", err)
	}

	if err := validators.GlobalEntitiesValidator.Struct(obj); err != nil {
		return nil, fmt.Errorf("failed to validate tool provider configuration: %w", err)
	}

	return &obj, nil
}
