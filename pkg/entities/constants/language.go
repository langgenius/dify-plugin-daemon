package constants

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type Language string

const (
	Python Language = "python"
	Go     Language = "go" // not supported yet
)

func isAvailableLanguage(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(Python):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("is_available_language", isAvailableLanguage)
}
