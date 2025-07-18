package models

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

type InstallTaskStatus string

const (
	InstallTaskStatusPending InstallTaskStatus = "pending"
	InstallTaskStatusRunning InstallTaskStatus = "running"
	InstallTaskStatusSuccess InstallTaskStatus = "success"
	InstallTaskStatusFailed  InstallTaskStatus = "failed"
)

type InstallTaskPluginStatus struct {
	PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifier"`
	Labels                 plugin_entities.I18nObject             `json:"labels"`
	Icon                   string                                 `json:"icon"`
	IconDark               string                                 `json:"icon_dark"`
	PluginID               string                                 `json:"plugin_id"`
	Status                 InstallTaskStatus                      `json:"status"`
	Message                string                                 `json:"message"`
}

type InstallTask struct {
	Model
	Status           InstallTaskStatus         `json:"status" gorm:"not null"`
	TenantID         string                    `json:"tenant_id" gorm:"type:uuid;not null"`
	TotalPlugins     int                       `json:"total_plugins" gorm:"not null"`
	CompletedPlugins int                       `json:"completed_plugins" gorm:"not null"`
	Plugins          []InstallTaskPluginStatus `json:"plugins" gorm:"serializer:json"`
}
