package models

import "time"

// HookID is a pointer to plugin id and tenant id, using it to identify the webhook plugin
type Webhook struct {
	Model
	HookID               string    `json:"hook_id" orm:"uniqueIndex;size:127;column:hook_id"`
	TenantID             string    `json:"tenant_id" orm:"index;size:64;column:tenant_id"`
	UserID               string    `json:"user_id" orm:"index;size:64;column:user_id"`
	PluginID             string    `json:"plugin_id" orm:"index;size:64;column:plugin_id"`
	ExpiredAt            time.Time `json:"expired_at" orm:"column:expired_at"`
	PluginInstallationId string    `json:"plugin_installation_id" orm:"index;size:64;column:plugin_installation_id"`
}
