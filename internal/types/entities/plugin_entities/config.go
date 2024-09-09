package plugin_entities

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type ConfigType string

const (
	CONFIG_TYPE_SECRET_INPUT   ConfigType = SECRET_INPUT
	CONFIG_TYPE_TEXT_INPUT     ConfigType = TEXT_INPUT
	CONFIG_TYPE_SELECT         ConfigType = SELECT
	CONFIG_TYPE_BOOLEAN        ConfigType = BOOLEAN
	CONFIG_TYPE_MODEL_SELECTOR ConfigType = MODEL_SELECTOR
	CONFIG_TYPE_APP_SELECTOR   ConfigType = APP_SELECTOR
	CONFIG_TYPE_TOOL_SELECTOR  ConfigType = TOOL_SELECTOR
)

type ModelConfigScope string

const (
	MODEL_CONFIG_SCOPE_ALL            ModelConfigScope = "all"
	MODEL_CONFIG_SCOPE_LLM            ModelConfigScope = "llm"
	MODEL_CONFIG_SCOPE_TEXT_EMBEDDING ModelConfigScope = "text-embedding"
	MODEL_CONFIG_SCOPE_RERANK         ModelConfigScope = "rerank"
	MODEL_CONFIG_SCOPE_TTS            ModelConfigScope = "tts"
	MODEL_CONFIG_SCOPE_SPEECH2TEXT    ModelConfigScope = "speech2text"
	MODEL_CONFIG_SCOPE_MODERATION     ModelConfigScope = "moderation"
	MODEL_CONFIG_SCOPE_VISION         ModelConfigScope = "vision"
)

type AppSelectorScope string

const (
	APP_SELECTOR_SCOPE_ALL        AppSelectorScope = "all"
	APP_SELECTOR_SCOPE_CHAT       AppSelectorScope = "chat"
	APP_SELECTOR_SCOPE_WORKFLOW   AppSelectorScope = "workflow"
	APP_SELECTOR_SCOPE_COMPLETION AppSelectorScope = "completion"
)

type ToolSelectorScope string

const (
	TOOL_SELECTOR_SCOPE_ALL      ToolSelectorScope = "all"
	TOOL_SELECTOR_SCOPE_PLUGIN   ToolSelectorScope = "plugin"
	TOOL_SELECTOR_SCOPE_API      ToolSelectorScope = "api"
	TOOL_SELECTOR_SCOPE_WORKFLOW ToolSelectorScope = "workflow"
)

func isCredentialType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(CONFIG_TYPE_SECRET_INPUT),
		string(CONFIG_TYPE_TEXT_INPUT),
		string(CONFIG_TYPE_SELECT),
		string(CONFIG_TYPE_BOOLEAN),
		string(CONFIG_TYPE_APP_SELECTOR):
		return true
	}
	return false
}

type ConfigOption struct {
	Value string     `json:"value" validate:"required"`
	Label I18nObject `json:"label" validate:"required"`
}

func isModelConfigScope(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(MODEL_CONFIG_SCOPE_LLM),
		string(MODEL_CONFIG_SCOPE_TEXT_EMBEDDING),
		string(MODEL_CONFIG_SCOPE_RERANK),
		string(MODEL_CONFIG_SCOPE_TTS),
		string(MODEL_CONFIG_SCOPE_SPEECH2TEXT),
		string(MODEL_CONFIG_SCOPE_MODERATION),
		string(MODEL_CONFIG_SCOPE_VISION):
		return true
	}
	return false
}

func isAppSelectorScope(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(APP_SELECTOR_SCOPE_ALL),
		string(APP_SELECTOR_SCOPE_CHAT),
		string(APP_SELECTOR_SCOPE_WORKFLOW),
		string(APP_SELECTOR_SCOPE_COMPLETION):
		return true
	}
	return false
}

func isToolSelectorScope(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(TOOL_SELECTOR_SCOPE_ALL),
		string(TOOL_SELECTOR_SCOPE_PLUGIN),
		string(TOOL_SELECTOR_SCOPE_API),
		string(TOOL_SELECTOR_SCOPE_WORKFLOW):
		return true
	}
	return false
}
func isScope(fl validator.FieldLevel) bool {
	// get parent and check if it's a provider config
	parent := fl.Parent().Interface()
	if provider_config, ok := parent.(ProviderConfig); ok {
		// check config type
		if provider_config.Type == CONFIG_TYPE_APP_SELECTOR {
			return isAppSelectorScope(fl)
		} else if provider_config.Type == CONFIG_TYPE_MODEL_SELECTOR {
			return isModelConfigScope(fl)
		} else if provider_config.Type == CONFIG_TYPE_TOOL_SELECTOR {
			return isToolSelectorScope(fl)
		} else {
			return false
		}
	}
	return false
}

func init() {
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")

	validators.GlobalEntitiesValidator.RegisterValidation("is_scope", isScope)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"is_scope",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("is_scope", "{0} is not a valid scope", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("is_scope", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("is_app_selector_scope", isAppSelectorScope)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"is_app_selector_scope",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("is_app_selector_scope", "{0} is not a valid app selector scope", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("is_app_selector_scope", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("is_model_config_scope", isModelConfigScope)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"is_model_config_scope",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("is_model_config_scope", "{0} is not a valid model config scope", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("is_model_config_scope", fe.Field())
			return t
		},
	)

	validators.GlobalEntitiesValidator.RegisterValidation("is_tool_selector_scope", isToolSelectorScope)
	validators.GlobalEntitiesValidator.RegisterTranslation(
		"is_tool_selector_scope",
		translator,
		func(ut ut.Translator) error {
			return ut.Add("is_tool_selector_scope", "{0} is not a valid tool selector scope", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("is_tool_selector_scope", fe.Field())
			return t
		},
	)
}

type ProviderConfig struct {
	Name        string         `json:"name" validate:"required,gt=0,lt=1024"`
	Type        ConfigType     `json:"type" validate:"required,credential_type"`
	Scope       *string        `json:"scope" validate:"omitempty,is_scope"`
	Required    bool           `json:"required"`
	Default     any            `json:"default" validate:"omitempty,is_basic_type"`
	Options     []ConfigOption `json:"options" validate:"omitempty,dive"`
	Label       I18nObject     `json:"label" validate:"required"`
	Helper      *I18nObject    `json:"helper" validate:"omitempty"`
	URL         *string        `json:"url" validate:"omitempty"`
	Placeholder *I18nObject    `json:"placeholder" validate:"omitempty"`
}

func init() {
	en := en.New()
	uni := ut.New(en, en)
	translator, _ := uni.GetTranslator("en")

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
}
