package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// automatically fetching a correct runtime according to the platform
func (p *PluginManager) GetPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginRuntimeSessionIOInterface, error) {
	if p.config.Platform == app.PLATFORM_SERVERLESS {
		return p.getServerlessPluginRuntime(pluginUniqueIdentifier)
	}

	return p.controlPanel.GetPluginRuntime(pluginUniqueIdentifier)
}

func (p *PluginManager) RemoveLocalPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) error {
	return p.controlPanel.RemoveLocalPlugin(pluginUniqueIdentifier)
}

// get local plugin runtime
func (p *PluginManager) GetLocalPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginRuntimeSessionIOInterface, error) {
	return p.controlPanel.GetPluginRuntime(pluginUniqueIdentifier)
}

// get serverless plugin runtime
func (p *PluginManager) GetServerlessPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginRuntimeSessionIOInterface, error) {
	// fetch serverless runtime
	return p.getServerlessPluginRuntime(pluginUniqueIdentifier)
}

func (p *PluginManager) ShutdownLocalPluginGracefully(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (<-chan error, error) {
	return p.controlPanel.ShutdownLocalPluginGracefully(pluginUniqueIdentifier)
}
