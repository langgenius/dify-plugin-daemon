package local_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

type LocalPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	entities.PluginRuntime

	wait_chan   chan bool
	io_identity string
}
