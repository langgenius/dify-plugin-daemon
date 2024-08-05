package plugin_entities

type InvokePluginPluginIdentity struct {
	PluginName    string `json:"plugin_name" binding:"required"`
	PluginVersion string `json:"plugin_version" binding:"required"`
}

type InvokePluginUserIdentity struct {
	TenantId string `json:"tenant_id" binding:"required"`
	UserId   string `json:"user_id" binding:"required"`
}

type InvokePluginRequest[T any] struct {
	InvokePluginPluginIdentity
	InvokePluginUserIdentity

	Data T `json:"data" binding:"required"`
}
