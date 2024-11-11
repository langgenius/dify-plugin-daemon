package local_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	plugin_entities.PluginRuntime

	waitChan    chan bool
	io_identity string

	// python interpreter path, currently only support python
	python_interpreter_path string

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	default_python_interpreter_path string

	waitChanLock      sync.Mutex
	wait_started_chan []chan bool
	waitStoppedChan   []chan bool

	isNotFirstStart bool
}

func NewLocalPluginRuntime(
	python_interpreter_path string,
) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		default_python_interpreter_path: python_interpreter_path,
	}
}
