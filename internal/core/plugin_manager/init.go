package plugin_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

var m sync.Map

func checkPluginExist(identity string) (entities.PluginRuntimeInterface, bool) {
	if v, ok := m.Load(identity); ok {
		return v.(entities.PluginRuntimeInterface), true
	}

	return nil, false
}
