package models

type TenantStorage struct {
	Model
	TenantID string `gorm:"column:tenant_id;type:varchar(255);not null;index"`
	PluginID string `gorm:"column:plugin_id;type:varchar(255);not null;index"`
	Size     int64  `gorm:"column:size;type:bigint;not null"`
}
