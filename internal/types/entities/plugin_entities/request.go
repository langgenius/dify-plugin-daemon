package plugin_entities

type InvokePluginUserIdentity struct {
	TenantId string `json:"tenant_id" validate:"required" uri:"tenant_id"`
	UserId   string `json:"user_id" validate:"required"`
}

type BasePluginIdentifier struct {
	PluginID string `json:"plugin_id"`
}

type InvokePluginRequest[T any] struct {
	InvokePluginUserIdentity
	BasePluginIdentifier

	UniqueIdentifier PluginUniqueIdentifier `json:"unique_identifier"`
	Data             T                      `json:"data" validate:"required"`
}
