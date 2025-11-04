package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type ControlPanelNotifier interface {
	// on local runtime starting
	OnLocalRuntimeStarting(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier)
	// on local runtime ready
	OnLocalRuntimeReady(runtime *local_runtime.LocalPluginRuntime)
	// on local runtime failed to start
	OnLocalRuntimeStartFailed(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier, err error)

	// on remote runtime connected
	OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime)
	// on remote runtime disconnected
	OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime)
}
