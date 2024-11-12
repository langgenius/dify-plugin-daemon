package local_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	plugin_entities.PluginRuntime

	waitChan   chan bool
	ioIdentity string

	// python interpreter path, currently only support python
	pythonInterpreterPath string

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	defaultPythonInterpreterPath string

	waitChanLock    sync.Mutex
	waitStartedChan []chan bool
	waitStoppedChan []chan bool

	isNotFirstStart bool
}

func NewLocalPluginRuntime(
	pythonInterpreterPath string,
) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		defaultPythonInterpreterPath: pythonInterpreterPath,
	}
}
