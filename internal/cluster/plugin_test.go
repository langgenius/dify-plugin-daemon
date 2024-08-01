package cluster

import (
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func getRandomPluginRuntime() entities.PluginRuntime {
	return entities.PluginRuntime{
		Config: plugin_entities.PluginDeclaration{
			Name:      uuid.New().String(),
			Version:   "0.0.1",
			Type:      plugin_entities.PluginType,
			Author:    "Yeuoly",
			CreatedAt: time.Now(),
			Plugins:   []string{"test"},
		},
	}
}
