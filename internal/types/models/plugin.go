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

func (p *Plugin) GetDeclaration() (*plugin_entities.PluginDeclaration, error) {
	declaration, err := parser.UnmarshalJson[plugin_entities.PluginDeclaration](p.Declaration)
	if err != nil {
		return nil, err
	}

	return &declaration, nil
}

type ServerlessRuntimeType string

const (
	SERVERLESS_RUNTIME_TYPE_AWS_LAMBDA ServerlessRuntimeType = "aws_lambda"
)

type ServerlessRuntime struct {
	Model
	PluginUniqueIdentifier string                `json:"plugin_unique_identifier" orm:"index;size:127"`
	FunctionURL            string                `json:"function_url" orm:"size:255"`
	FunctionName           string                `json:"function_name" orm:"size:127"`
	Type                   ServerlessRuntimeType `json:"type" orm:"size:127"`
	Declaration            string                `json:"declaration" orm:"type:text;size:65535"`
	Checksum               string                `json:"checksum" orm:"size:127"`
}

func (p *ServerlessRuntime) GetDeclaration() (*plugin_entities.PluginDeclaration, error) {
	declaration, err := parser.UnmarshalJson[plugin_entities.PluginDeclaration](p.Declaration)
	if err != nil {
		return nil, err
	}

	return &declaration, nil
}

func (p *ServerlessRuntime) SetDeclaration(declaration *plugin_entities.PluginDeclaration) {
	p.Declaration = parser.MarshalJson(declaration)
}
