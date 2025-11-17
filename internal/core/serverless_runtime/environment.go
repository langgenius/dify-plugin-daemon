package serverless_runtime

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (r *ServerlessPluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	checksum, err := r.Checksum()
	if err != nil {
		return "", err
	}
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", r.Config.Identity(), checksum))
}
