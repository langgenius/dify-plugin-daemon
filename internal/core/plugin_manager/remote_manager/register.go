package remote_manager

import "github.com/langgenius/dify-plugin-daemon/internal/service/install_service"

func (plugin *RemotePluginRuntime) Register() error {
	_, installation, err := install_service.InstallPlugin(
		plugin.tenant_id, "", plugin, "remote", map[string]any{},
	)
	if err != nil {
		return err
	}
	plugin.installation_id = installation.ID
	return nil
}

func (plugin *RemotePluginRuntime) Unregister() error {
	identity, err := plugin.Identity()
	if err != nil {
		return err
	}
	return install_service.UninstallPlugin(
		plugin.tenant_id,
		plugin.installation_id,
		identity,
		plugin.Type(),
	)
}
