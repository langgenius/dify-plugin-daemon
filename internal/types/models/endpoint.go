package models

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// HookID is a pointer to plugin id and tenant id, using it to identify the endpoint plugin
type Endpoint struct {
	Model
	Name        string                                       `json:"name" gorm:"size:127;column:name;default:'default'"`
	HookID      string                                       `json:"hook_id" gorm:"unique;size:127;column:hook_id"`
	TenantID    string                                       `json:"tenant_id" gorm:"index;size:64;column:tenant_id"`
	UserID      string                                       `json:"user_id" gorm:"index;size:64;column:user_id"`
	PluginID    string                                       `json:"plugin_id" gorm:"index;size:64;column:plugin_id"`
	ExpiredAt   time.Time                                    `json:"expired_at" gorm:"column:expired_at"`
	Enabled     bool                                         `json:"enabled" gorm:"column:enabled"`
	Settings    map[string]any                               `json:"settings" gorm:"column:settings;serializer:json"`
	Declaration *plugin_entities.EndpointProviderDeclaration `json:"declaration" gorm:"-"` // not stored in db
}
