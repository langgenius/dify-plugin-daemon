package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (m *PluginManager) checkPluginExist(identity string) (plugin_entities.PluginRuntimeInterface, bool) {
	if v, ok := m.m.Load(identity); ok {
		return v.(plugin_entities.PluginRuntimeInterface), true
	}

	return nil, false
}
