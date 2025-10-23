package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// RequestLaunchPlugin requests the control panel to launch a plugin
// This function only trig the control panel to launch plugin
func RequestLaunchPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	runtimeType plugin_entities.PluginRuntimeType,
	pluginManager *plugin_manager.PluginManager,
	config *app.Config,
) (*stream.Stream[ControlPanelSignalEntity], error) {
	// TODO: implement this
	// 1. trig the daemon to launch plugin and subscribe the monitoring
	return nil, nil
}
