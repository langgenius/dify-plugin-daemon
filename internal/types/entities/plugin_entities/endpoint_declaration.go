package plugin_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type EndpointMethod string

const (
	EndpointMethodHead    EndpointMethod = "HEAD"
	EndpointMethodGet     EndpointMethod = "GET"
	EndpointMethodPost    EndpointMethod = "POST"
	EndpointMethodPut     EndpointMethod = "PUT"
	EndpointMethodDelete  EndpointMethod = "DELETE"
	EndpointMethodOptions EndpointMethod = "OPTIONS"
)

func isAvailableMethod(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	switch method {
	case string(EndpointMethodHead),
		string(EndpointMethodGet),
		string(EndpointMethodPost),
		string(EndpointMethodPut),
		string(EndpointMethodDelete),
		string(EndpointMethodOptions):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("is_available_endpoint_method", isAvailableMethod)
}

type EndpointDeclaration struct {
	Path   string         `json:"path" yaml:"path" validate:"required"`
	Method EndpointMethod `json:"method" yaml:"method" validate:"required,is_available_endpoint_method"`
}

type EndpointProviderDeclaration struct {
	Settings map[string]ProviderConfig `json:"settings" validate:"omitempty,dive"`
}
