package local_manager

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities"

type LocalPluginRuntime struct {
	entities.PluginRuntime

	io_identity string
}
