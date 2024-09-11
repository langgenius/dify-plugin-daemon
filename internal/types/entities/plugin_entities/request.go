package plugin_entities

type InvokePluginUserIdentity struct {
	TenantId string `json:"tenant_id" binding:"required"`
	UserId   string `json:"user_id" binding:"required"`
}

type BasePluginIdentifier struct {
	PluginUniqueIdentifier PluginUniqueIdentifier `json:"plugin_unique_identifier"`
}

type InvokePluginRequest[T any] struct {
	InvokePluginUserIdentity
	BasePluginIdentifier

	Data T `json:"data" binding:"required"`
}
