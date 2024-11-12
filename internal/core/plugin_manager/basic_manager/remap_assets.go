package basic_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

// RemapAssets will take the assets and remap them to a media id
func (r *BasicPluginRuntime) RemapAssets(
	declaration *plugin_entities.PluginDeclaration,
	assets map[string][]byte,
) error {
	assetsIds, err := r.mediaManager.RemapAssets(declaration, assets)
	if err != nil {
		return err
	}

	r.assetsIds = assetsIds
	return nil
}
