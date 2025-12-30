package types

import (
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

type EnvConfig struct {
	InnerAPIURL string `json:"inner_api_url" validate:"required"`
	InnerAPIKey string `json:"inner_api_key" validate:"required"`
	TenantID    string `json:"tenant_id"`
	UserID      string `json:"user_id"`
}

type DifyToolIdentity struct {
	Author   string                     `json:"author"`
	Name     string                     `json:"name"`
	Label    plugin_entities.I18nObject `json:"label"`
	Provider string                     `json:"provider"`
}

type DifyToolDeclaration struct {
	ProviderType   requests.ToolType                `json:"provider_type" yaml:"provider_type" validate:"required"`
	Identity       DifyToolIdentity                 `json:"identity" yaml:"identity" validate:"required"`
	Description    plugin_entities.ToolDescription  `json:"description" yaml:"description" validate:"required"`
	Parameters     []plugin_entities.ToolParameter  `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	OutputSchema   plugin_entities.ToolOutputSchema `json:"output_schema,omitempty" yaml:"output_schema,omitempty"`
	CredentialType string                           `json:"credential_type" yaml:"credential_type" validate:"omitempty"`
	CredentialId   string                           `json:"credential_id" yaml:"credential_id" validate:"omitempty"`
}

type DifyConfig struct {
	Env   EnvConfig             `json:"env"`
	Tools []DifyToolDeclaration `json:"tools"`
}
