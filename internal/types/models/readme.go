package models

type PluginReadme struct {
	Model
	TenantID               string `json:"tenant_id" gorm:"column:tenant_id;type:uuid;index;not null"`
	PluginUniqueIdentifier string `json:"plugin_unique_identifier" gorm:"column:plugin_unique_identifier;size:255;index;not null"`
	Language               string `json:"language" gorm:"column:language;size:10;not null"`
	Content                string `json:"content" gorm:"column:content;type:text;not null"`
}
