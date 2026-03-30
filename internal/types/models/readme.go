package models

// PluginReadmeRecord represents the readme content of a plugin in a specific language.
//
// NOTE: originally, this was named PluginReadme, but `tenant_id` column was added, which is a technical issue
// as readme is not a tenant-specific resource.
// however, it's hard to only delete a column in production environments
// without affecting existing data, so, a new model PluginReadmeRecord is introduced to replace PluginReadme without
// tenant specificity.
type PluginReadmeRecord struct {
	Model
	PluginUniqueIdentifier string `json:"plugin_unique_identifier" gorm:"column:plugin_unique_identifier;size:255;index;not null"`
	Language               string `json:"language" gorm:"column:language;size:10;not null"`
	Content                string `json:"content" gorm:"column:content;type:text;not null"`
}
