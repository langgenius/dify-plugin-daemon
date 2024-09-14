package remote_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	// copy a new declaration
	config := r.Config
	config.Author = r.tenant_id
	checksum, _ := r.Checksum()
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", config.Identity(), checksum))
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}
