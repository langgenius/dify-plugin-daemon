package plugin_entities

type InvokePluginUserIdentity struct {
	TenantId string `json:"tenant_id" validate:"required"`
	UserId   string `json:"user_id" validate:"required"`
}

type BasePluginIdentifier struct {
	PluginUniqueIdentifier PluginUniqueIdentifier `json:"plugin_unique_identifier"`
}

type InvokePluginRequest[T any] struct {
	InvokePluginUserIdentity
	BasePluginIdentifier

	Data T `json:"data" validate:"required"`
}
