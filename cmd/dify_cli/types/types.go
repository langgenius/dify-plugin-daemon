package types

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

type EnvConfig struct {
	InnerAPIURL string `json:"inner_api_url"`
	InnerAPIKey string `json:"inner_api_key"`
}

type DifyConfig struct {
	Env       EnvConfig                               `json:"env"`
	Providers []plugin_entities.ToolProviderDeclaration `json:"providers"`
}
