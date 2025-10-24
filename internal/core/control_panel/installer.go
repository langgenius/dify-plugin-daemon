package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (c *ControlPanel) InstallPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[PluginInstallResponse], error) {
	// TODO: implement this
	return nil, nil
}
