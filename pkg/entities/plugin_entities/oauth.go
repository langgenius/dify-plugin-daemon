package plugin_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type SupportedAuthType string

const (
	SupportedAuthTypeOAuth2 SupportedAuthType = "oauth"
	SupportedAuthTypeAPIKey SupportedAuthType = "api_key"
)

func isSupportedAuthType(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case string(SupportedAuthTypeOAuth2), string(SupportedAuthTypeAPIKey):
		return true
	default:
		return false
	}
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("supported_auth_type", isSupportedAuthType)
}

type OAuthSchema struct {
	// ClientSchema contains client_id, client_secret, redirect_uri, etc. which are required to be set by system admin
	ClientSchema []ProviderConfig `json:"client_schema" yaml:"client_schema" validate:"omitempty,dive"`

	// CredentialsSchema contains schema of access_token, refresh_token, etc.
	CredentialsSchema []ProviderConfig `json:"credentials_schema" yaml:"credentials_schema" validate:"omitempty,dive"`
}
