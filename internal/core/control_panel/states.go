package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type pluginInstanceUnit struct {
	pluginInstance *local_runtime.LocalPluginRuntime
	state          *local_runtime.LocalPluginMonitoringStatus
}

type ControllerStates struct {
	// plugin instance mapping, for managing and monitoring plugin instances
	pluginInstanceMapping mapping.Map[
		plugin_entities.PluginUniqueIdentifier, pluginInstanceUnit,
	]
}
