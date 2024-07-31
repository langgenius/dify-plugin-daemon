package models

type PluginInstallationStatus string

type PluginInstallation struct {
	Model
	TenantID string `json:"tenant_id" orm:"index;type:uuid;"`
	UserID   string `json:"user_id" orm:"index;type:uuid;"`
	PluginID string `json:"plugin_id" orm:"index;size:127"`
}
