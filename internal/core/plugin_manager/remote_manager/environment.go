package remote_manager

import (
	"fmt"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	identity := strings.Join([]string{r.tenant_id, r.Configuration().Identity()}, "/")
	checksum, _ := r.Checksum()
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", identity, checksum))
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}
