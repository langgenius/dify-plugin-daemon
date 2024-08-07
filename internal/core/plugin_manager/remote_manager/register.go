package remote_manager

import "github.com/langgenius/dify-plugin-daemon/internal/service/install_service"

func (plugin *RemotePluginRuntime) Register() error {
	installation_id, err := install_service.InstallPlugin(plugin.tenant_id, "", plugin, map[string]any{})
	if err != nil {
		return err
	}
	plugin.installation_id = installation_id
	return nil
}

func (plugin *RemotePluginRuntime) Unregister() error {
	return install_service.UninstallPlugin(plugin.tenant_id, plugin.installation_id, plugin)
}
