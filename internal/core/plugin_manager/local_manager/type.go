package local_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	plugin_entities.PluginRuntime

	wait_chan   chan bool
	io_identity string

	// python interpreter path, currently only support python
	python_interpreter_path         string
	default_python_interpreter_path string

	wait_chan_lock    sync.Mutex
	wait_started_chan []chan bool
	wait_stopped_chan []chan bool
}

func NewLocalPluginRuntime(
	python_interpreter_path string,
) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		default_python_interpreter_path: python_interpreter_path,
	}
}
