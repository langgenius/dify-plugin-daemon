package plugin_entities

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type ToolIdentity struct {
	Author string     `json:"author" validate:"required"`
	Name   string     `json:"name" validate:"required"`
	Label  I18nObject `json:"label" validate:"required"`
}

type ToolParameterOption struct {
	Value string     `json:"value" validate:"required"`
	Label I18nObject `json:"label" validate:"required"`
}

type ToolParameterType string

const (
	TOOL_PARAMETER_TYPE_STRING       ToolParameterType = "string"
	TOOL_PARAMETER_TYPE_NUMBER       ToolParameterType = "number"
	TOOL_PARAMETER_TYPE_BOOLEAN      ToolParameterType = "boolean"
	TOOL_PARAMETER_TYPE_SELECT       ToolParameterType = "select"
	TOOL_PARAMETER_TYPE_SECRET_INPUT ToolParameterType = "secret_input"
	TOOL_PARAMETER_TYPE_FILE         ToolParameterType = "file"
)

func isToolParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(TOOL_PARAMETER_TYPE_STRING),
		string(TOOL_PARAMETER_TYPE_NUMBER),
		string(TOOL_PARAMETER_TYPE_BOOLEAN),
		string(TOOL_PARAMETER_TYPE_SELECT),
		string(TOOL_PARAMETER_TYPE_SECRET_INPUT),
		string(TOOL_PARAMETER_TYPE_FILE):
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

type ToolParameter struct {
	Name             string                `json:"name" validate:"required,gt=0,lt=1024"`
	Label            I18nObject            `json:"label" validate:"required"`
	HumanDescription I18nObject            `json:"human_description" validate:"required"`
	Type             ToolParameterType     `json:"type" validate:"required,tool_parameter_type"`
	Form             ToolParameterForm     `json:"form" validate:"required,tool_parameter_form"`
	LLMDescription   string                `json:"llm_description" validate:"omitempty"`
	Required         bool                  `json:"required" validate:"required"`
	Default          any                   `json:"default" validate:"omitempty,is_basic_type"`
	Min              *float64              `json:"min" validate:"omitempty"`
	Max              *float64              `json:"max" validate:"omitempty"`
	Options          []ToolParameterOption `json:"options" validate:"omitempty,dive"`
}

type ToolDescription struct {
	Human I18nObject `json:"human" validate:"required"`
	LLM   string     `json:"llm" validate:"required"`
}

type ToolConfiguration struct {
	Identity    ToolIdentity    `json:"identity" validate:"required"`
	Description ToolDescription `json:"description" validate:"required"`
	Parameters  []ToolParameter `json:"parameters" validate:"omitempty,dive"`
}

type ToolCredentialsOption struct {
	Value string     `json:"value" validate:"required"`
	Label I18nObject `json:"label" validate:"required"`
}

type CredentialType string

const (
	CREDENTIAL_TYPE_SECRET_INPUT CredentialType = "secret-input"
	CREDENTIAL_TYPE_TEXT_INPUT   CredentialType = "text-input"
	CREDENTIAL_TYPE_SELECT       CredentialType = "select"
	CREDENTIAL_TYPE_BOOLEAN      CredentialType = "boolean"
)

func isCredentialType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(CREDENTIAL_TYPE_SECRET_INPUT),
		string(CREDENTIAL_TYPE_TEXT_INPUT),
		string(CREDENTIAL_TYPE_SELECT),
		string(CREDENTIAL_TYPE_BOOLEAN):
		return true
	}
	return false
}

type ToolProviderCredential struct {
	Name        string                  `json:"name" validate:"required,gt=0,lt=1024"`
	Type        CredentialType          `json:"type" validate:"required,credential_type"`
	Required    bool                    `json:"required"`
	Default     any                     `json:"default" validate:"omitempty,is_basic_type"`
	Options     []ToolCredentialsOption `json:"options" validate:"omitempty,dive"`
	Label       I18nObject              `json:"label" validate:"required"`
	Helper      *I18nObject             `json:"helper" validate:"omitempty"`
	URL         *string                 `json:"url" validate:"omitempty"`
	Placeholder *I18nObject             `json:"placeholder" validate:"omitempty"`
}

type ToolLabel string

const (
	TOOL_LABEL_SEARCH        ToolLabel = "search"
	TOOL_LABEL_IMAGE         ToolLabel = "image"
	TOOL_LABEL_VIDEOS        ToolLabel = "videos"
	TOOL_LABEL_WEATHER       ToolLabel = "weather"
	TOOL_LABEL_FINANCE       ToolLabel = "finance"
	TOOL_LABEL_DESIGN        ToolLabel = "design"
	TOOL_LABEL_TRAVEL        ToolLabel = "travel"
	TOOL_LABEL_SOCIAL        ToolLabel = "social"
	TOOL_LABEL_NEWS          ToolLabel = "news"
	TOOL_LABEL_MEDICAL       ToolLabel = "medical"
	TOOL_LABEL_PRODUCTIVITY  ToolLabel = "productivity"
	TOOL_LABEL_EDUCATION     ToolLabel = "education"
	TOOL_LABEL_BUSINESS      ToolLabel = "business"
	TOOL_LABEL_ENTERTAINMENT ToolLabel = "entertainment"
	TOOL_LABEL_UTILITIES     ToolLabel = "utilities"
	TOOL_LABEL_OTHER         ToolLabel = "other"
)

func isToolLabel(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(TOOL_LABEL_SEARCH),
		string(TOOL_LABEL_IMAGE),
		string(TOOL_LABEL_VIDEOS),
		string(TOOL_LABEL_WEATHER),
		string(TOOL_LABEL_FINANCE),
		string(TOOL_LABEL_DESIGN),
		string(TOOL_LABEL_TRAVEL),
		string(TOOL_LABEL_SOCIAL),
		string(TOOL_LABEL_NEWS),
		string(TOOL_LABEL_MEDICAL),
		string(TOOL_LABEL_PRODUCTIVITY),
		string(TOOL_LABEL_EDUCATION),
		string(TOOL_LABEL_BUSINESS),
		string(TOOL_LABEL_ENTERTAINMENT),
		string(TOOL_LABEL_UTILITIES),
		string(TOOL_LABEL_OTHER):
		return true
	}
	return false
}

type ToolProviderIdentity struct {
	Author      string      `json:"author" validate:"required"`
	Name        string      `json:"name" validate:"required"`
	Description I18nObject  `json:"description" validate:"required"`
	Icon        []byte      `json:"icon" validate:"required"`
	Label       I18nObject  `json:"label" validate:"required"`
	Tags        []ToolLabel `json:"tags" validate:"required,dive,tool_label"`
}

type ToolProviderConfiguration struct {
	Identity          ToolProviderIdentity              `json:"identity" validate:"required"`
	CredentialsSchema map[string]ToolProviderCredential `json:"credentials_schema" validate:"omitempty,dive"`
	Tools             []ToolConfiguration               `json:"tools" validate:"required,dive"`
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

	validators.GlobalEntitiesValidator.RegisterValidation("credential_type", isCredentialType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"credential_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("credential_type", "{0} is not a valid credential type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("credential_type", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("tool_label", isToolLabel)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"tool_label",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("tool_label", "{0} is not a valid tool label", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("tool_label", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("is_basic_type", isGenericType)
}

func UnmarshalToolProviderConfiguration(data []byte) (*ToolProviderConfiguration, error) {
	obj, err := parser.UnmarshalJsonBytes[ToolProviderConfiguration](data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool provider configuration: %w", err)
	}

	if err := validators.GlobalEntitiesValidator.Struct(obj); err != nil {
		return nil, fmt.Errorf("failed to validate tool provider configuration: %w", err)
	}

	return &obj, nil
}
