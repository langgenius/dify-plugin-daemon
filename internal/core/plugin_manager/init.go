package plugin_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

var m sync.Map

func checkPluginExist(name string) (*entities.PluginRuntime, bool) {
	if v, ok := m.Load(name); ok {
		if plugin, ok := v.(*entities.PluginRuntime); ok {
			return plugin, true
		}
	}

	return nil, false
}
