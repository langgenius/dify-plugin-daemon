package models

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type Plugin struct {
	Model
	// PluginUniqueIdentifier is a unique identifier for the plugin, it contains version and checksum
	PluginUniqueIdentifier string `json:"plugin_unique_identifier" orm:"index;size:127"`
	// PluginID is the id of the plugin, only plugin name is considered
	PluginID     string                            `json:"id" orm:"index;size:127"`
	Refers       int                               `json:"refers" orm:"default:0"`
	InstallType  plugin_entities.PluginRuntimeType `json:"install_type" orm:"size:127;index"`
	ManifestType plugin_entities.DifyManifestType  `json:"manifest_type" orm:"size:127"`
	Declaration  string                            `json:"declaration" orm:"type:text;size:65535"`
}

func (p *Plugin) GetDeclaration() (plugin_entities.PluginDeclaration, error) {
	return parser.UnmarshalJson[plugin_entities.PluginDeclaration](p.Declaration)
}
