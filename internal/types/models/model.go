package models

type AIModelInstallation struct {
	Model
	Provider               string `json:"provider" gorm:"column:provider;size:127;index;not null"`
	TenantID               string `json:"tenant_id" gorm:"column:tenant_id;type:uuid;index;not null"`
	PluginUniqueIdentifier string `json:"plugin_unique_identifier" gorm:"index;size:255"`
	PluginID               string `json:"plugin_id" gorm:"index;size:255"`
}
