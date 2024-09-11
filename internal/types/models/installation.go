package models

import (
	"encoding/json"
)

type PluginInstallationStatus string

type PluginInstallation struct {
	Model
	TenantID               string `json:"tenant_id" gorm:"index;type:uuid;"`
	PluginID               string `json:"plugin_id" gorm:"index;size:127"`
	PluginUniqueIdentifier string `json:"plugin_unique_identifier" gorm:"index;size:127"`
	Config                 string `json:"config"`
}

func (p *PluginInstallation) ConfigMap() (map[string]any, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(p.Config), &config); err != nil {
		return nil, err
	}
	return config, nil
}
