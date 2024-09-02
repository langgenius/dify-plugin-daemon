package plugin_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type ProviderType string

const (
	PROVIDER_TYPE_MODEL    ProviderType = "model"
	PROVIDER_TYPE_TOOL     ProviderType = "tool"
	PROVIDER_TYPE_ENDPOINT ProviderType = "endpoint"
)

func isAvailableProviderType(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	return str == string(PROVIDER_TYPE_MODEL) || str == string(PROVIDER_TYPE_TOOL) || str == string(PROVIDER_TYPE_ENDPOINT)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("provider_type", isAvailableProviderType)
}

type GenericProviderDeclaration struct {
	Type     ProviderType   `json:"type" yaml:"type" validate:"required,provider_type"`
	Provider map[string]any `json:"provider" yaml:"provider" validate:"required"`
}
