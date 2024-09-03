package remote_manager

import (
	"fmt"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginIdentity, error) {
	identity := strings.Join([]string{r.Configuration().Identity(), r.tenant_id}, ":")
	return plugin_entities.PluginIdentity(fmt.Sprintf("%s@%s", identity, r.Checksum())), nil
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}
