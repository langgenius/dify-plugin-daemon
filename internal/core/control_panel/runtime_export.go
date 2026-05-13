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
	if runtime, ok := c.debuggingPluginRuntime.Load(pluginUniqueIdentifier); ok {
		return runtime, nil
	}

	if runtime, ok := c.localPluginRuntimes.Load(pluginUniqueIdentifier); ok {
		return runtime, nil
	}

	return nil, ErrPluginRuntimeNotFound
}
