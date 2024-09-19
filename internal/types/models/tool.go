package models

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"

type ToolInstallation struct {
	Model
	TenantID               string                                  `json:"tenant_id" gorm:"column:tenant_id;type:uuid;index;not null"`
	Provider               string                                  `json:"provider" gorm:"column:provider;size:127;index;not null"`
	PluginUniqueIdentifier string                                  `json:"plugin_unique_identifier" gorm:"index;size:127"`
	PluginID               string                                  `json:"plugin_id" gorm:"index;size:127"`
	Declaration            plugin_entities.ToolProviderDeclaration `json:"declaration" gorm:"serializer:json;type:text;size:65535;not null"`
}