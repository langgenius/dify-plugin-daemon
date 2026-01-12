package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// GetPluginRuntime returns the plugin runtime for the given plugin unique identifier
// it automatically detects the runtime type and returns the corresponding runtime
//
// NOTE: serverless runtime is not supported in this method
// it only works for runtime which actually running on this machine
func (c *ControlPanel) GetPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginRuntimeSessionIOInterface, error) {
	if pluginUniqueIdentifier.RemoteLike() {
		runtime, ok := c.debuggingPluginRuntime.Load(pluginUniqueIdentifier)
		if !ok {
			return nil, ErrPluginRuntimeNotFound
		}
		return runtime, nil
	} else {
		runtime, ok := c.localPluginRuntimes.Load(pluginUniqueIdentifier)
		if !ok {
			return nil, ErrPluginRuntimeNotFound
		}
		return runtime, nil
	}
}

func (c *ControlPanel) GetLocalPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*local_runtime.LocalPluginRuntime, bool) {
	return c.localPluginRuntimes.Load(pluginUniqueIdentifier)
}

func (c *ControlPanel) GetDebuggingPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*debugging_runtime.RemotePluginRuntime, bool) {
	return c.debuggingPluginRuntime.Load(pluginUniqueIdentifier)
}
