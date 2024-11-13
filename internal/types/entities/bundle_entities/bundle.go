package bundle_entities

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
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
}
