package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// UninstallFromLocal uninstalls a plugin from local storage
// once deleted, local runtime will automatically shutdown and exit after several time
func (p *PluginManager) UninstallFromLocal(identity plugin_entities.PluginUniqueIdentifier) error {
	if err := p.installedBucket.Delete(identity); err != nil {
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
