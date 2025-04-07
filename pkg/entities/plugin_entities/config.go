package plugin_entities

import (
	"errors"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type ConfigType string

const (
	CONFIG_TYPE_SECRET_INPUT   ConfigType = SECRET_INPUT
	CONFIG_TYPE_TEXT_INPUT     ConfigType = TEXT_INPUT
	CONFIG_TYPE_SELECT         ConfigType = SELECT
	CONFIG_TYPE_BOOLEAN        ConfigType = BOOLEAN
	CONFIG_TYPE_MODEL_SELECTOR ConfigType = MODEL_SELECTOR
	CONFIG_TYPE_APP_SELECTOR   ConfigType = APP_SELECTOR
	// CONFIG_TYPE_TOOL_SELECTOR  ConfigType = TOOL_SELECTOR
	CONFIG_TYPE_TOOLS_SELECTOR ConfigType = TOOLS_SELECTOR
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
	MODEL_CONFIG_SCOPE_DOCUMENT       ModelConfigScope = "document"
	MODEL_CONFIG_SCOPE_TOOL_CALL      ModelConfigScope = "tool-call"
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
		string(CONFIG_TYPE_APP_SELECTOR),
		string(CONFIG_TYPE_MODEL_SELECTOR),
		string(CONFIG_TYPE_TOOLS_SELECTOR):
		return true
	}
	return false
}

type ConfigOption struct {
	Value string     `json:"value" validate:"required,lt=128"`
	Label I18nObject `json:"label" validate:"required"`
}

func isModelConfigScope(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// split by and symbol
	scopes := strings.Split(value, "&")
	for _, scope := range scopes {
		// trim space
		scope = strings.TrimSpace(scope)
		switch scope {
		case string(MODEL_CONFIG_SCOPE_LLM),
			string(MODEL_CONFIG_SCOPE_TEXT_EMBEDDING),
			string(MODEL_CONFIG_SCOPE_RERANK),
			string(MODEL_CONFIG_SCOPE_TTS),
			string(MODEL_CONFIG_SCOPE_SPEECH2TEXT),
			string(MODEL_CONFIG_SCOPE_MODERATION),
			string(MODEL_CONFIG_SCOPE_VISION),
			string(MODEL_CONFIG_SCOPE_DOCUMENT),
			string(MODEL_CONFIG_SCOPE_TOOL_CALL):
			return true
		}
	}
	return false
}

func isAppSelectorScope(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// split by and symbol
	scopes := strings.Split(value, "&")
	for _, scope := range scopes {
		// trim space
		scope = strings.TrimSpace(scope)
		switch scope {
		case string(APP_SELECTOR_SCOPE_ALL),
			string(APP_SELECTOR_SCOPE_CHAT),
			string(APP_SELECTOR_SCOPE_WORKFLOW),
			string(APP_SELECTOR_SCOPE_COMPLETION):
			return true
		}
	}
	return false
}

func isToolSelectorScope(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// split by and symbol
	scopes := strings.Split(value, "&")
	for _, scope := range scopes {
		// trim space
		scope = strings.TrimSpace(scope)
		switch scope {
		case string(TOOL_SELECTOR_SCOPE_ALL),
			string(TOOL_SELECTOR_SCOPE_PLUGIN),
			string(TOOL_SELECTOR_SCOPE_API),
			string(TOOL_SELECTOR_SCOPE_WORKFLOW):
			return true
		}
	}
	return false
}
func isScope(fl validator.FieldLevel) bool {
	// get parent and check if it's a provider config
	parent := fl.Parent().Interface()
	if providerConfig, ok := parent.(ProviderConfig); ok {
		// check config type
		if providerConfig.Type == CONFIG_TYPE_APP_SELECTOR {
			return isAppSelectorScope(fl)
		} else if providerConfig.Type == CONFIG_TYPE_MODEL_SELECTOR {
			return isModelConfigScope(fl)
		} else {
			return false
		}

		//else if providerConfig.Type == CONFIG_TYPE_TOOL_SELECTOR {
		//return isToolSelectorScope(fl)
		//}
	}
	if toolParameter, ok := parent.(ToolParameter); ok {
		if toolParameter.Type == TOOL_PARAMETER_TYPE_APP_SELECTOR {
			return isAppSelectorScope(fl)
		} else if toolParameter.Type == TOOL_PARAMETER_TYPE_MODEL_SELECTOR {
			return isModelConfigScope(fl)
		} else {
			return false
		}

		// else if toolParameter.Type == TOOL_PARAMETER_TYPE_TOOL_SELECTOR {
		// 	return isToolSelectorScope(fl)
		// }
	}

	if agentStrategyParameter, ok := parent.(AgentStrategyParameter); ok {
		if agentStrategyParameter.Type == AGENT_STRATEGY_PARAMETER_TYPE_APP_SELECTOR {
			return isAppSelectorScope(fl)
		} else if agentStrategyParameter.Type == AGENT_STRATEGY_PARAMETER_TYPE_MODEL_SELECTOR {
			return isModelConfigScope(fl)
		} else {
			return false
		}

		//else if agentStrategyParameter.Type == AGENT_STRATEGY_PARAMETER_TYPE_TOOLS_SELECTOR {
		//	return isToolSelectorScope(fl)
		//}
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
	Name        string         `json:"name" validate:"omitempty,gt=0,lt=1024"`
	Type        ConfigType     `json:"type" validate:"required,credential_type"`
	Scope       *string        `json:"scope" validate:"omitempty,is_scope"`
	Required    bool           `json:"required"`
	Default     any            `json:"default" validate:"omitempty,is_basic_type"`
	Options     []ConfigOption `json:"options" validate:"omitempty,lt=128,dive"`
	Label       I18nObject     `json:"label" validate:"required"`
	Help        *I18nObject    `json:"help" validate:"omitempty"`
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

// ValidateProviderConfigs validates the provider configs
func ValidateProviderConfigs(settings map[string]any, configs []ProviderConfig) error {
	if len(settings) > 64 {
		return errors.New("too many setting fields")
	}

	configsMap := make(map[string]ProviderConfig)
	for _, config := range configs {
		configsMap[config.Name] = config
	}

	for config_name, config := range configsMap {
		v, ok := settings[config_name]
		if (!ok || v == nil) && config.Required {
			return errors.New("missing required setting: " + config_name)
		}

		if !ok || v == nil {
			continue
		}

		// check type
		switch config.Type {
		case CONFIG_TYPE_TEXT_INPUT:
			if _, ok := v.(string); !ok {
				return errors.New("setting " + config_name + " is not a string")
			}
		case CONFIG_TYPE_SECRET_INPUT:
			if _, ok := v.(string); !ok {
				return errors.New("setting " + config_name + " is not a string")
			}
		case CONFIG_TYPE_SELECT:
			if _, ok := v.(string); !ok {
				return errors.New("setting " + config_name + " is not a string")
			}
			// check if value is in options
			found := false
			for _, option := range config.Options {
				if v == option.Value {
					found = true
					break
				}
			}
			if !found {
				return errors.New("setting " + config_name + " is not a valid option")
			}
		case CONFIG_TYPE_BOOLEAN:
			if _, ok := v.(bool); !ok {
				return errors.New("setting " + config_name + " is not a boolean")
			}
		case CONFIG_TYPE_APP_SELECTOR:
			m, ok := v.(map[string]any)
			if !ok {
				return errors.New("setting " + config_name + " is not a map")
			}
			// check keys
			if _, ok := m["app_id"]; !ok {
				return errors.New("setting " + config_name + " is missing app_id")
			}
		case CONFIG_TYPE_MODEL_SELECTOR:
			m, ok := v.(map[string]any)
			if !ok {
				return errors.New("setting " + config_name + " is not a map")
			}
			// check keys
			if _, ok := m["provider"]; !ok {
				return errors.New("setting " + config_name + " is missing provider")
			}
			if _, ok := m["model"]; !ok {
				return errors.New("setting " + config_name + " is missing model")
			}
			if _, ok := m["model_type"]; !ok {
				return errors.New("setting " + config_name + " is missing model_type")
			}
			// check scope
			if config.Scope != nil {
				switch *config.Scope {
				case string(MODEL_CONFIG_SCOPE_ALL):
					// do nothing
				case string(MODEL_CONFIG_SCOPE_LLM):
					// do nothing
				case string(MODEL_CONFIG_SCOPE_TEXT_EMBEDDING):
					// do nothing
				case string(MODEL_CONFIG_SCOPE_RERANK):
					// score_threshold, top_n
					if _, ok := m["score_threshold"]; !ok {
						return errors.New("setting " + config_name + " is missing score_threshold")
					}
					if _, ok := m["top_n"]; !ok {
						return errors.New("setting " + config_name + " is missing top_n")
					}
				case string(MODEL_CONFIG_SCOPE_TTS):
					// voice
					if _, ok := m["voice"]; !ok {
						return errors.New("setting " + config_name + " is missing voice")
					}
				case string(MODEL_CONFIG_SCOPE_SPEECH2TEXT):
					// do nothing
				case string(MODEL_CONFIG_SCOPE_MODERATION):
					// do nothing
				case string(MODEL_CONFIG_SCOPE_VISION):
					// the same as llm
					if _, ok := m["completion_params"]; !ok {
						return errors.New("setting " + config_name + " is missing completion_params")
					}
				default:
					return errors.New("setting " + config_name + " is not a valid model config scope")
				}
			}
			// case CONFIG_TYPE_TOOL_SELECTOR:
			// 	m, ok := v.(map[string]any)
			// 	if !ok {
			// 		return errors.New("setting " + config_name + " is not a map")
			// 	}
			// 	// check keys
			// 	if _, ok := m["provider"]; !ok {
			// 		return errors.New("setting " + config_name + " is missing provider")
			// 	}
			// 	if _, ok := m["tool"]; !ok {
			// 		return errors.New("setting " + config_name + " is missing tool")
			// 	}
			// 	if _, ok := m["tool_type"]; !ok {
			// 		return errors.New("setting " + config_name + " is missing tool_type")
			// 	}
		}
	}

	return nil
}
