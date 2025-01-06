package debugging_runtime

import "github.com/langgenius/dify-plugin-daemon/internal/service/install_service"

func (plugin *RemotePluginRuntime) Register() error {
	_, installation, err := install_service.InstallPlugin(
		plugin.tenantId, "", plugin, "remote", map[string]any{},
	)
	if err != nil {
		return err
	}
	plugin.installationId = installation.ID
	return nil
}

func (plugin *RemotePluginRuntime) Unregister() error {
	identity, err := plugin.Identity()
	if err != nil {
		return err
	}
	return install_service.UninstallPlugin(
		plugin.tenantId,
		plugin.installationId,
		identity,
		plugin.Type(),
	)
}
