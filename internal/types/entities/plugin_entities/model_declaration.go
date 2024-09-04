package plugin_entities

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
	"github.com/shopspring/decimal"
)

type ModelType string

const (
	MODEL_TYPE_LLM            ModelType = "llm"
	MODEL_TYPE_TEXT_EMBEDDING ModelType = "text-embedding"
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
	CONFIGURATE_METHOD_PREDEFINED_MODEL   ModelProviderConfigurateMethod = "predefined-model"
	CONFIGURATE_METHOD_CUSTOMIZABLE_MODEL ModelProviderConfigurateMethod = "customizable-model"
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
	Name        string              `json:"name" validate:"required,lt=256"`
	UseTemplate *string             `json:"use_template" validate:"omitempty,lt=256"`
	Label       *I18nObject         `json:"label" validate:"omitempty"`
	Type        *ModelParameterType `json:"type" validate:"omitempty,model_parameter_type"`
	Help        *I18nObject         `json:"help" validate:"omitempty"`
	Required    bool                `json:"required"`
	Default     *any                `json:"default" validate:"omitempty,is_basic_type"`
	Min         *float64            `json:"min" validate:"omitempty"`
	Max         *float64            `json:"max" validate:"omitempty"`
	Precision   *int                `json:"precision" validate:"omitempty"`
	Options     []string            `json:"options" validate:"omitempty,dive,lt=256"`
}

func isParameterRule(fl validator.FieldLevel) bool {
	// if use_template is empty, then label, type should be required
	// try get the value of use_template
	use_template_handle := fl.Field().FieldByName("UseTemplate")
	// check if use_template is null pointer
	if use_template_handle.IsNil() {
		// label and type should be required
		// try get the value of label
		if fl.Field().FieldByName("Label").IsNil() {
			return false
		}

		// try get the value of type
		if fl.Field().FieldByName("Type").IsNil() {
			return false
		}
	}

	return true
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

type ModelDeclaration struct {
	Model           string                         `json:"model" validate:"required,lt=256"`
	Label           I18nObject                     `json:"label" validate:"required"`
	ModelType       ModelType                      `json:"model_type" validate:"required,model_type"`
	Features        []string                       `json:"features" validate:"omitempty,lte=256,dive,lt=256"`
	FetchFrom       ModelProviderConfigurateMethod `json:"fetch_from" validate:"omitempty,model_provider_configurate_method"`
	ModelProperties map[string]any                 `json:"model_properties" validate:"omitempty,dive,is_basic_type"`
	Deprecated      bool                           `json:"deprecated"`
	ParameterRules  []ModelParameterRule           `json:"parameter_rules" validate:"omitempty,lte=128,dive,parameter_rule"`
	PriceConfig     *ModelPriceConfig              `json:"price_config" validate:"omitempty"`
}

type ModelProviderFormType string

const (
	FORM_TYPE_TEXT_INPUT   ModelProviderFormType = "text-input"
	FORM_TYPE_SECRET_INPUT ModelProviderFormType = "secret-input"
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
	ShowOn []ModelProviderFormShowOnObject `json:"show_on" validate:"omitempty,lte=16,dive"`
}

type ModelProviderCredentialFormSchema struct {
	Variable    string                          `json:"variable" validate:"required,lt=256"`
	Label       I18nObject                      `json:"label" validate:"required"`
	Type        ModelProviderFormType           `json:"type" validate:"required,model_provider_form_type"`
	Required    bool                            `json:"required"`
	Default     *string                         `json:"default" validate:"omitempty,lt=256"`
	Options     []ModelProviderFormOption       `json:"options" validate:"omitempty,lte=128,dive"`
	Placeholder *I18nObject                     `json:"placeholder" validate:"omitempty"`
	MaxLength   int                             `json:"max_length"`
	ShowOn      []ModelProviderFormShowOnObject `json:"show_on" validate:"omitempty,lte=16,dive"`
}

type ModelProviderCredentialSchema struct {
	CredentialFormSchemas []ModelProviderCredentialFormSchema `json:"credential_form_schemas" validate:"omitempty,lte=32,dive"`
}

type FieldModelSchema struct {
	Label       I18nObject  `json:"label" validate:"required"`
	Placeholder *I18nObject `json:"placeholder" validate:"omitempty"`
}

type ModelCredentialSchema struct {
	Model                  FieldModelSchema                    `json:"model" validate:"required"`
	CredentialsFormSchemas []ModelProviderCredentialFormSchema `json:"credentials_form_schemas" validate:"omitempty,lte=32,dive"`
}

type ModelProviderHelpEntity struct {
	Title I18nObject `json:"title" validate:"required"`
	URL   I18nObject `json:"url" validate:"required"`
}

type ModelProviderDeclaration struct {
	Provider                 string                           `json:"provider" yaml:"provider" validate:"required,lt=256"`
	Label                    I18nObject                       `json:"label" yaml:"label" validate:"required"`
	Description              *I18nObject                      `json:"description" yaml:"description,omitempty" validate:"omitempty"`
	IconSmall                *I18nObject                      `json:"icon_small" yaml:"icon_small,omitempty" validate:"omitempty"`
	IconLarge                *I18nObject                      `json:"icon_large" yaml:"icon_large,omitempty" validate:"omitempty"`
	Background               *string                          `json:"background" yaml:"background,omitempty" validate:"omitempty"`
	Help                     *ModelProviderHelpEntity         `json:"help" yaml:"help,omitempty" validate:"omitempty"`
	SupportedModelTypes      []ModelType                      `json:"supported_model_types" yaml:"supported_model_types" validate:"required,lte=16,dive,model_type"`
	ConfigurateMethods       []ModelProviderConfigurateMethod `json:"configurate_methods" yaml:"configurate_methods" validate:"required,lte=16,dive,model_provider_configurate_method"`
	Models                   []string                         `json:"models" yaml:"models" validate:"required,lte=1024"`
	ProviderCredentialSchema *ModelProviderCredentialSchema   `json:"provider_credential_schema" yaml:"provider_credential_schema,omitempty" validate:"omitempty"`
	ModelCredentialSchema    *ModelCredentialSchema           `json:"model_credential_schema" yaml:"model_credential_schema,omitempty" validate:"omitempty"`
	ModelDeclarations        []ModelDeclaration               `json:"model_declarations" yaml:"model_declarations,omitempty"`
}

func init() {
	// init validator
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")
	// register translations for default validators
	en_translations.RegisterDefaultTranslations(validators.GlobalEntitiesValidator, translator)

	validators.GlobalEntitiesValidator.RegisterValidation("model_type", isModelType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
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

	validators.GlobalEntitiesValidator.RegisterValidation("model_provider_configurate_method", isModelProviderConfigurateMethod)
	validators.GlobalEntitiesValidator.RegisterTranslation(
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

	validators.GlobalEntitiesValidator.RegisterValidation("model_provider_form_type", isModelProviderFormType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
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

	validators.GlobalEntitiesValidator.RegisterValidation("model_parameter_type", isModelParameterType)
	validators.GlobalEntitiesValidator.RegisterTranslation(
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

	validators.GlobalEntitiesValidator.RegisterValidation("parameter_rule", isParameterRule)

	validators.GlobalEntitiesValidator.RegisterValidation("is_basic_type", isBasicType)
}
