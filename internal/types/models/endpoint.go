package models

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

// HookID is a pointer to plugin id and tenant id, using it to identify the endpoint plugin
type Endpoint struct {
	Model
	HookID      string                                       `json:"hook_id" gorm:"unique;size:127;column:hook_id"`
	TenantID    string                                       `json:"tenant_id" gorm:"index;size:64;column:tenant_id"`
	UserID      string                                       `json:"user_id" gorm:"index;size:64;column:user_id"`
	PluginID    string                                       `json:"plugin_id" gorm:"index;size:64;column:plugin_id"`
	ExpiredAt   time.Time                                    `json:"expired_at" gorm:"column:expired_at"`
	Enabled     bool                                         `json:"enabled" gorm:"column:enabled"`
	Settings    string                                       `json:"settings" gorm:"column:settings;size:2048"`
	Declaration *plugin_entities.EndpointProviderDeclaration `json:"declaration" gorm:"-"` // not stored in db
}

func (e *Endpoint) GetSettings() map[string]any {
	d, _ := parser.UnmarshalJson2Map(e.Settings)
	return d
}

func (e *Endpoint) SetSettings(settings map[string]any) {
	e.Settings = parser.MarshalJson(settings)
}
