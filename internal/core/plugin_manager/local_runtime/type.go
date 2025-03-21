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

	// python env init timeout
	pythonEnvInitTimeout int

	// python compileall extra args
	pythonCompileAllExtraArgs string

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	defaultPythonInterpreterPath string

	pipMirrorUrl    string
	pipPreferBinary bool
	pipVerbose      bool
	pipExtraArgs    string

	// proxy settings
	HttpProxy  string
	HttpsProxy string

	waitChanLock    sync.Mutex
	waitStartedChan []chan bool
	waitStoppedChan []chan bool

	isNotFirstStart bool
}

type LocalPluginRuntimeConfig struct {
	PythonInterpreterPath     string
	PythonEnvInitTimeout      int
	PythonCompileAllExtraArgs string
	HttpProxy                 string
	HttpsProxy                string
	PipMirrorUrl              string
	PipPreferBinary           bool
	PipVerbose                bool
	PipExtraArgs              string
}

func NewLocalPluginRuntime(config LocalPluginRuntimeConfig) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		defaultPythonInterpreterPath: config.PythonInterpreterPath,
		pythonEnvInitTimeout:         config.PythonEnvInitTimeout,
		pythonCompileAllExtraArgs:    config.PythonCompileAllExtraArgs,
		HttpProxy:                    config.HttpProxy,
		HttpsProxy:                   config.HttpsProxy,
		pipMirrorUrl:                 config.PipMirrorUrl,
		pipPreferBinary:              config.PipPreferBinary,
		pipVerbose:                   config.PipVerbose,
		pipExtraArgs:                 config.PipExtraArgs,
	}
}
