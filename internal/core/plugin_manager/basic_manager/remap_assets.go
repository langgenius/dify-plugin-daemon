package basic_manager

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"

// RemapAssets will take the assets and remap them to a media id
func (r *BasicPluginRuntime) RemapAssets(
	declaration *plugin_entities.PluginDeclaration,
	assets map[string][]byte,
) error {
	// TODO: implement
	return nil
}
