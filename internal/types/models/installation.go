package models

type PluginInstallationStatus string

type PluginInstallation struct {
	Model
	TenantID               string `json:"tenant_id" gorm:"index;type:uuid;"`
	PluginID               string `json:"plugin_id" gorm:"index;size:127"`
	PluginUniqueIdentifier string `json:"plugin_unique_identifier" gorm:"index;size:127"`
	RuntimeType            string `json:"runtime_type" gorm:"size:127"`
	EndpointsSetups        int    `json:"endpoints_setups"`
	EndpointsActive        int    `json:"endpoints_active"`
}
