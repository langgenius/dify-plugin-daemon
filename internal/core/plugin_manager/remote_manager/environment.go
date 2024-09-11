package remote_manager

import (
	"fmt"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	identity := strings.Join([]string{r.Configuration().Identity(), r.tenant_id}, ":")
	checksum, _ := r.Checksum()
	return plugin_entities.PluginUniqueIdentifier(fmt.Sprintf("%s@%s", identity, checksum)), nil
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}
