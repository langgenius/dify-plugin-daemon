package models

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

type Plugin struct {
	Model
	PluginID          string                           `json:"id" orm:"index;size:127"`
	ConfigurationText string                           `json:"configuration_text" orm:"type:text"`
	Refers            int                              `json:"refers" orm:"default:0"`
	Checksum          string                           `json:"checksum" orm:"size:127"`
	InstallType       entities.PluginRuntimeType       `json:"install_type" orm:"size:127"`
	ManifestType      plugin_entities.DifyManifestType `json:"manifest_type" orm:"size:127"`
}
