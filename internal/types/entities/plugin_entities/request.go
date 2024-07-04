package plugin_entities

type InvokePluginRequestData interface {
	InvokeToolRequest | InvokeModelRequest
}

type InvokeModelRequest struct {
}

type InvokePluginRequest[T InvokePluginRequestData] struct {
	PluginName    string `json:"plugin_name" binding:"required"`
	PluginVersion string `json:"plugin_version" binding:"required"`
	TenantId      string `json:"tenant_id" binding:"required"`
	Data          T      `json:"data" binding:"required"`
}

type InvokeToolRequest struct {
	ProviderName string `json:"provider_name" binding:"required"`
	ToolName     string `json:"tool_name" binding:"required"`
	ToolRuntime  struct {
	} `json:"tool_runtime" binding:"required"`
	Parameters map[string]interface{} `json:"parameters" binding:"required"`
}
