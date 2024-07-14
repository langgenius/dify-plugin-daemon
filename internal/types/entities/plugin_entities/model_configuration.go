package plugin_entities

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/shopspring/decimal"
)

type ModelType string

const (
	MODEL_TYPE_LLM            ModelType = "llm"
	MODEL_TYPE_TEXT_EMBEDDING ModelType = "text_embedding"
	MODEL_TYPE_RERANKING      ModelType = "rerank"
	MODEL_TYPE_SPEECH2TEXT    ModelType = "speech2text"
	MODEL_TYPE_MODERATION     ModelType = "moderation"
	MODEL_TYPE_TTS            ModelType = "tts"
	MODEL_TYPE_TEXT2IMG       ModelType = "text2img"
)

func isModelType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(MODEL_TYPE_LLM),
		string(MODEL_TYPE_TEXT_EMBEDDING),
		string(MODEL_TYPE_RERANKING),
		string(MODEL_TYPE_SPEECH2TEXT),
		string(MODEL_TYPE_MODERATION),
		string(MODEL_TYPE_TTS),
		string(MODEL_TYPE_TEXT2IMG):
		return true
	}
	return false
}

type ModelProviderConfigurateMethod string

const (
	CONFIGURATE_METHOD_PREDEFINED_MODEL   ModelProviderConfigurateMethod = "predefined_model"
	CONFIGURATE_METHOD_CUSTOMIZABLE_MODEL ModelProviderConfigurateMethod = "customizable_model"
)

func isModelProviderConfigurateMethod(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(CONFIGURATE_METHOD_PREDEFINED_MODEL),
		string(CONFIGURATE_METHOD_CUSTOMIZABLE_MODEL):
		return true
	}
	return false
}

type ModelParameterType string

const (
	PARAMETER_TYPE_FLOAT   ModelParameterType = "float"
	PARAMETER_TYPE_INT     ModelParameterType = "int"
	PARAMETER_TYPE_STRING  ModelParameterType = "string"
	PARAMETER_TYPE_BOOLEAN ModelParameterType = "boolean"
)

func isModelParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PARAMETER_TYPE_FLOAT),
		string(PARAMETER_TYPE_INT),
		string(PARAMETER_TYPE_STRING),
		string(PARAMETER_TYPE_BOOLEAN):
		return true
	}
	return false
}

type ModelParameterRule struct {
	Name        string             `json:"name" validate:"required,lt=256"`
	UseTemplate *string            `json:"use_template" validate:"omitempty,lt=256"`
	Label       I18nObject         `json:"label" validate:"required"`
	Type        ModelParameterType `json:"type" validate:"required,model_parameter_type"`
	Help        *I18nObject        `json:"help" validate:"omitempty"`
	Required    bool               `json:"required"`
	Default     *any               `json:"default" validate:"omitempty,is_basic_type"`
	Min         *float64           `json:"min" validate:"omitempty"`
	Max         *float64           `json:"max" validate:"omitempty"`
	Precision   *int               `json:"precision" validate:"omitempty"`
	Options     []string           `json:"options" validate:"omitempty,dive,lt=256"`
}

type ModelPriceConfig struct {
	Input    decimal.Decimal  `json:"input" validate:"required"`
	Output   *decimal.Decimal `json:"output" validate:"omitempty"`
	Unit     decimal.Decimal  `json:"unit" validate:"required"`
	Currency string           `json:"currency" validate:"required"`
}

type ModelPricing struct {
	PricePerUnit float64 `json:"price_per_unit" validate:"required"`
}

type ModelConfiguration struct {
	Model           string                         `json:"model" validate:"required,lt=256"`
	Label           I18nObject                     `json:"label" validate:"required"`
	ModelType       ModelType                      `json:"model_type" validate:"required,model_type"`
	Features        []string                       `json:"features" validate:"omitempty,dive,lt=256"`
	FetchFrom       ModelProviderConfigurateMethod `json:"fetch_from" validate:"required,model_provider_configurate_method"`
	ModelProperties map[string]any                 `json:"model_properties" validate:"omitempty,dive,is_basic_type"`
	Deprecated      bool                           `json:"deprecated"`
	ParameterRules  []ModelParameterRule           `json:"parameter_rules" validate:"omitempty,dive"`
	PriceConfig     *ModelPriceConfig              `json:"price_config" validate:"omitempty"`
}

type ModelProviderFormType string

const (
	FORM_TYPE_TEXT_INPUT   ModelProviderFormType = "text_input"
	FORM_TYPE_SECRET_INPUT ModelProviderFormType = "secret_input"
	FORM_TYPE_SELECT       ModelProviderFormType = "select"
	FORM_TYPE_RADIO        ModelProviderFormType = "radio"
	FORM_TYPE_SWITCH       ModelProviderFormType = "switch"
)

func isModelProviderFormType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(FORM_TYPE_TEXT_INPUT),
		string(FORM_TYPE_SECRET_INPUT),
		string(FORM_TYPE_SELECT),
		string(FORM_TYPE_RADIO),
		string(FORM_TYPE_SWITCH):
		return true
	}
	return false
}

type ModelProviderFormShowOnObject struct {
	Variable string `json:"variable" validate:"required,lt=256"`
	Value    string `json:"value" validate:"required,lt=256"`
}

type ModelProviderFormOption struct {
	Label  I18nObject                      `json:"label" validate:"required"`
	Value  string                          `json:"value" validate:"required,lt=256"`
	ShowOn []ModelProviderFormShowOnObject `json:"show_on" validate:"omitempty,dive,lt=16"`
}

type ModelProviderCredentialFormSchema struct {
	Variable    string                          `json:"variable" validate:"required,lt=256"`
	Label       I18nObject                      `json:"label" validate:"required"`
	Type        ModelProviderFormType           `json:"type" validate:"required,model_provider_form_type"`
	Required    bool                            `json:"required"`
	Default     *string                         `json:"default" validate:"omitempty,lt=256"`
	Options     []ModelProviderFormOption       `json:"options" validate:"omitempty,dive,lt=128"`
	Placeholder *I18nObject                     `json:"placeholder" validate:"omitempty"`
	MaxLength   int                             `json:"max_length"`
	ShowOn      []ModelProviderFormShowOnObject `json:"show_on" validate:"omitempty,dive,lt=16"`
}

type ModelProviderCredentialSchema struct {
	CredentialFormSchemas []ModelProviderCredentialFormSchema `json:"credential_form_schemas" validate:"omitempty,dive,lt=32"`
}

type FieldModelSchema struct {
	Label       I18nObject  `json:"label" validate:"required"`
	Placeholder *I18nObject `json:"placeholder" validate:"omitempty"`
}

type ModelCredentialSchema struct {
	Model                  FieldModelSchema                    `json:"model" validate:"required"`
	CredentialsFormSchemas []ModelProviderCredentialFormSchema `json:"credentials_form_schemas" validate:"omitempty,dive,lt=32"`
}

type ModelProviderHelpEntity struct {
	Title I18nObject `json:"title" validate:"required,lt=256"`
	URL   string     `json:"url" validate:"required,lt=256"`
}

type ModelProviderConfiguration struct {
	Provider                 string                           `json:"provider" validate:"required,lt=256"`
	Label                    I18nObject                       `json:"label" validate:"required"`
	Description              *I18nObject                      `json:"description" validate:"omitempty"`
	IconSmall                *I18nObject                      `json:"icon_small" validate:"omitempty"`
	IconLarge                *I18nObject                      `json:"icon_large" validate:"omitempty"`
	Background               *string                          `json:"background" validate:"omitempty"`
	Help                     *ModelProviderHelpEntity         `json:"help" validate:"omitempty"`
	SupportedModelTypes      []ModelType                      `json:"supported_model_types" validate:"required,dive,model_type,unique"`
	ConfigurateMethods       []ModelProviderConfigurateMethod `json:"configurate_methods" validate:"required,dive,model_provider_configurate_method,unique"`
	Models                   []ModelConfiguration             `json:"models" validate:"omitempty,dive,lt=1024"`
	ProviderCredentialSchema *ModelProviderCredentialSchema   `json:"provider_credential_schema" validate:"omitempty"`
	ModelCredentialSchema    *ModelCredentialSchema           `json:"model_credential_schema" validate:"omitempty"`
}

var (
	global_model_provider_validator = validator.New()
)

func init() {
	// init validator
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")
	// register translations for default validators
	en_translations.RegisterDefaultTranslations(global_model_provider_validator, translator)

	global_model_provider_validator.RegisterValidation("model_type", isModelType)
	global_model_provider_validator.RegisterTranslation(
		"model_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_type", "{0} is not a valid model type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_type", fe.Field())
			return t
		},
	)

	global_model_provider_validator.RegisterValidation("model_provider_configurate_method", isModelProviderConfigurateMethod)
	global_model_provider_validator.RegisterTranslation(
		"model_provider_configurate_method",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_provider_configurate_method", "{0} is not a valid model provider configurate method", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_provider_configurate_method", fe.Field())
			return t
		},
	)

	global_model_provider_validator.RegisterValidation("model_provider_form_type", isModelProviderFormType)
	global_model_provider_validator.RegisterTranslation(
		"model_provider_form_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_provider_form_type", "{0} is not a valid model provider form type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_provider_form_type", fe.Field())
			return t
		},
	)

	global_model_provider_validator.RegisterValidation("model_parameter_type", isModelParameterType)
	global_model_provider_validator.RegisterTranslation(
		"model_parameter_type",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("model_parameter_type", "{0} is not a valid model parameter type", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("model_parameter_type", fe.Field())
			return t
		},
	)

	global_model_provider_validator.RegisterValidation("is_basic_type", isGenericType)
}
