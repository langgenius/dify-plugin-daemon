package remote_manager

import "github.com/langgenius/dify-plugin-daemon/internal/service/install_service"

func (plugin *RemotePluginRuntime) Register() error {
	return install_service.InstallPlugin(plugin.tenant_id, "", plugin, map[string]any{})
}
