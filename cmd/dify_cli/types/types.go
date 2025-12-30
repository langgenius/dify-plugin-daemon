package types

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

type EnvConfig struct {
	InnerAPIURL string `json:"inner_api_url"`
	InnerAPIKey string `json:"inner_api_key"`
	TenantID    string `json:"tenant_id"`
	UserID      string `json:"user_id"`
}

type DifyConfig struct {
	Env       EnvConfig                               `json:"env"`
	Providers []plugin_entities.ToolProviderDeclaration `json:"providers"`
}
