package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// RequestLaunchPlugin requests the control panel to launch a plugin
// This function only trig the control panel to launch plugin
func (c *ControlPanel) RequestLaunchLocalPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[any], error) {

	return nil, nil
}
