package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

func (m *PluginManager) checkPluginExist(identity string) (entities.PluginRuntimeInterface, bool) {
	if v, ok := m.m.Load(identity); ok {
		return v.(entities.PluginRuntimeInterface), true
	}

	return nil, false
}
