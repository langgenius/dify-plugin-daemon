package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func SubscribePluginMonitoring(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[ControlPanelSignalEntity], error) {
	// TODO: implement this
	// 1. get the plugin instance from the plugin instance mapping
	// 2. subscribe the monitoring
	// 3. return the stream
	return nil, nil
}
