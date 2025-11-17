package plugin_manager

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

func (p *PluginManager) GetPluginRuntime(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginRuntimeSessionIOInterface, error) {
	return p.controlPanel.GetPluginRuntime(pluginUniqueIdentifier)
}

func (p *PluginManager) RemoveLocalPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) error {
	return p.controlPanel.RemoveLocalPlugin(pluginUniqueIdentifier)
}

func (p *PluginManager) ShutdownLocalPluginGracefully(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (<-chan error, error) {
	return p.controlPanel.ShutdownLocalPluginGracefully(pluginUniqueIdentifier)
}
