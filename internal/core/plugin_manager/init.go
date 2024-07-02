package plugin_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

var m sync.Map

func checkPluginExist(identity string) (*entities.PluginRuntime, bool) {
	if v, ok := m.Load(identity); ok {
		return v.(*entities.PluginRuntime), true
	}

	return nil, false
}
