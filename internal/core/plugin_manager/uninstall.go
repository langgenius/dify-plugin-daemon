package plugin_manager

import (
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

// UninstallFromLocal uninstalls a plugin from local storage
// once deleted, local runtime will automatically shutdown and exit after several time
func (p *PluginManager) UninstallFromLocal(identity plugin_entities.PluginUniqueIdentifier) error {
	plugin_installation_path := filepath.Join(p.pluginStoragePath, identity.String())
	if err := os.RemoveAll(plugin_installation_path); err != nil {
		return err
	}
	// send shutdown runtime
	runtime, ok := p.m.Load(identity.String())
	if !ok {
		// no runtime to shutdown, already uninstalled
		return nil
	}
	runtime.Stop()
	return nil
}
