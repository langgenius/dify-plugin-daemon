package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// RequestLaunchPlugin requests the control panel to launch a plugin
// This function only trig the control panel to launch plugin
func (c *ControlPanel) RequestLaunchPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	runtimeType plugin_entities.PluginRuntimeType,
) (*stream.Stream[any], error) {
	// TODO: implement this
	// 1. trig the daemon to launch plugin and subscribe the monitoring
	return nil, nil
}
