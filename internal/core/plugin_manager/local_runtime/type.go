package local_runtime

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	basic_runtime.BasicChecksum
	plugin_entities.PluginRuntime

	waitChan   chan bool
	ioIdentity string

	// python interpreter path, currently only support python
	pythonInterpreterPath string

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	defaultPythonInterpreterPath string

	// proxy settings
	proxyHttp  string
	proxyHttps string

	waitChanLock    sync.Mutex
	waitStartedChan []chan bool
	waitStoppedChan []chan bool

	isNotFirstStart bool
}

func NewLocalPluginRuntime(
	pythonInterpreterPath string,
	proxyHttp string,
	proxyHttps string,
) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		defaultPythonInterpreterPath: pythonInterpreterPath,
		proxyHttp:                    proxyHttp,
		proxyHttps:                   proxyHttps,
	}
}
