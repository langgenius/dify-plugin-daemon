package bundle_entities

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"gopkg.in/yaml.v3"
)

type Bundle struct {
	Name         string                             `json:"name" yaml:"name" validate:"required"`
	Labels       plugin_entities.I18nObject         `json:"labels" yaml:"labels" validate:"required"`
	Description  plugin_entities.I18nObject         `json:"description" yaml:"description" validate:"required"`
	Icon         string                             `json:"icon" yaml:"icon" validate:"required"`
	Version      manifest_entities.Version          `json:"version" yaml:"version" validate:"required,version"`
	Author       string                             `json:"author" yaml:"author" validate:"required"`
	Type         manifest_entities.DifyManifestType `json:"type" yaml:"type" validate:"required,eq=bundle"`
	Dependencies []Dependency                       `json:"dependencies" yaml:"dependencies" validate:"required"`
	Tags         []manifest_entities.PluginTag      `json:"tags" yaml:"tags" validate:"omitempty,dive,plugin_tag,max=128"`
}

// for api, avoid pydantic validation error
func (b *Bundle) MarshalJSON() ([]byte, error) {
	type alias Bundle
	p := alias(*b)

	if p.Tags == nil {
		p.Tags = []manifest_entities.PluginTag{}
	}

	return json.Marshal(p)
}

// for unmarshal yaml
func (b *Bundle) UnmarshalYAML(node *yaml.Node) error {
	// avoid nil tags
	type alias Bundle

	p := &struct {
		*alias `yaml:",inline"`
	}{
		alias: (*alias)(b),
	}

	if err := node.Decode(p); err != nil {
		return err
	}

	if p.Tags == nil {
		p.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}
