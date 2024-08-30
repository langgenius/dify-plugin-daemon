package models

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

// HookID is a pointer to plugin id and tenant id, using it to identify the endpoint plugin
type Endpoint struct {
	Model
	HookID               string    `json:"hook_id" orm:"uniqueIndex;size:127;column:hook_id"`
	TenantID             string    `json:"tenant_id" orm:"index;size:64;column:tenant_id"`
	UserID               string    `json:"user_id" orm:"index;size:64;column:user_id"`
	PluginID             string    `json:"plugin_id" orm:"index;size:64;column:plugin_id"`
	ExpiredAt            time.Time `json:"expired_at" orm:"column:expired_at"`
	Settings             string    `json:"settings" orm:"column:settings;size:2048"`
	PluginInstallationId string    `json:"plugin_installation_id" orm:"index;size:64;column:plugin_installation_id"`
}

func (e *Endpoint) GetSettings() map[string]any {
	d, _ := parser.UnmarshalJson2Map(e.Settings)
	return d
}
