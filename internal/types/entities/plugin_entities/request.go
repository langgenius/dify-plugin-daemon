package plugin_entities

type InvokePluginRequest[T any] struct {
	PluginName    string `json:"plugin_name" binding:"required"`
	PluginVersion string `json:"plugin_version" binding:"required"`
	TenantId      string `json:"tenant_id" binding:"required"`
	UserId        string `json:"user_id" binding:"required"`
	Data          T      `json:"data" binding:"required"`
}
