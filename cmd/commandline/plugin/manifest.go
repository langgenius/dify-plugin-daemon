package plugin

import (
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// as for some reasons, fields like `custom_setup_enabled` cannot be serialized into the manifest,
// so we need to use a new struct to represent the manifest with extra fields.
type ManifestWithExtra struct {
	plugin_entities.PluginDeclaration
	customSetupEnabled bool
}
