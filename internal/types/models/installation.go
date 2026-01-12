package models

type PluginInstallationStatus string

type PluginInstallation struct {
	Model
	TenantID               string         `json:"tenant_id" gorm:"index;uniqueIndex:idx_tenant_plugin;type:uuid;"`
	PluginID               string         `json:"plugin_id" gorm:"index;uniqueIndex:idx_tenant_plugin;size:255"`
	PluginUniqueIdentifier string         `json:"plugin_unique_identifier" gorm:"index;size:255"`
	RuntimeType            string         `json:"runtime_type" gorm:"size:127"`
	EndpointsSetups        int            `json:"endpoints_setups"`
	EndpointsActive        int            `json:"endpoints_active"`
	Source                 string         `json:"source" gorm:"column:source;size:63"`
	Meta                   map[string]any `json:"meta" gorm:"column:meta;serializer:json"`
	UpgradeState           string         `json:"upgrade_state" gorm:"size:63;index"`
	UpgradeOriginalVersion string         `json:"upgrade_original_version" gorm:"size:255"`
	UpgradeStartedAt       *int64         `json:"upgrade_started_at"`
}
