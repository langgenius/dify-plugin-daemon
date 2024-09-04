package local_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	plugin_entities.PluginRuntime

	wait_chan   chan bool
	io_identity string

	// python interpreter path, currently only support python
	python_interpreter_path string
}
