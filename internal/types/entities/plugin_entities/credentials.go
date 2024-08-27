package plugin_entities

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type ConfigType string

const (
	CONFIG_TYPE_SECRET_INPUT      ConfigType = SECRET_INPUT
	CONFIG_TYPE_TEXT_INPUT        ConfigType = TEXT_INPUT
	CONFIG_TYPE_SELECT            ConfigType = SELECT
	CONFIG_TYPE_BOOLEAN           ConfigType = BOOLEAN
	CONFIG_TYPE_CHAT_APP_ID       ConfigType = CHAT_APP_ID
	CONFIG_TYPE_COMPLETION_APP_ID ConfigType = COMPLETION_APP_ID
	CONFIG_TYPE_WORKFLOW_APP_ID   ConfigType = WORKFLOW_APP_ID
	CONFIG_TYPE_MODEL_CONFIG      ConfigType = MODEL_CONFIG
)

func isCredentialType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(CONFIG_TYPE_SECRET_INPUT),
		string(CONFIG_TYPE_TEXT_INPUT),
		string(CONFIG_TYPE_SELECT),
		string(CONFIG_TYPE_BOOLEAN),
		string(CONFIG_TYPE_CHAT_APP_ID),
		string(CONFIG_TYPE_COMPLETION_APP_ID),
		string(CONFIG_TYPE_WORKFLOW_APP_ID):
		return true
	}
	return false
}

type ConfigOption struct {
	Value string     `json:"value" validate:"required"`
	Label I18nObject `json:"label" validate:"required"`
}

type ProviderConfig struct {
	Name        string         `json:"name" validate:"required,gt=0,lt=1024"`
	Type        ConfigType     `json:"type" validate:"required,credential_type"`
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
