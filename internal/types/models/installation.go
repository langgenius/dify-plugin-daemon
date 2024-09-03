package models

import "encoding/json"

type PluginInstallationStatus string

type PluginInstallation struct {
	Model
	TenantID       string `json:"tenant_id" orm:"index;type:uuid;"`
	UserID         string `json:"user_id" orm:"index;type:uuid;"`
	PluginID       string `json:"plugin_id" orm:"index;size:127"`
	PluginIdentity string `json:"plugin_identity" orm:"index;size:127"`
	Config         string `json:"config"`
}

func (p *PluginInstallation) ConfigMap() (map[string]any, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(p.Config), &config); err != nil {
		return nil, err
	}
	return config, nil
}
