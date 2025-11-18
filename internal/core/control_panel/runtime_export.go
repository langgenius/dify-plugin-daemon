package controlpanel

import (
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
