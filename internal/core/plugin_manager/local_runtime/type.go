package local_runtime

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	basic_runtime.BasicChecksum
	plugin_entities.PluginRuntime

	waitChan chan bool

	// python interpreter path, currently only support python
	pythonInterpreterPath string

	// python env init timeout
	pythonEnvInitTimeout int

	// python compileall extra args
	pythonCompileAllExtraArgs string

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	defaultPythonInterpreterPath string
	uvPath                       string

	pipMirrorUrl    string
	pipPreferBinary bool
	pipVerbose      bool
	pipExtraArgs    string

	// proxy settings
	HttpProxy  string
	HttpsProxy string
	NoProxy    string

	waitChanLock    sync.Mutex
	waitStartedChan []chan bool
	waitStoppedChan []chan bool

	stdoutBufferSize    int
	stdoutMaxBufferSize int

	isNotFirstStart bool

	stdioHolder *stdioHolder
}

type LocalPluginRuntimeConfig struct {
	PythonInterpreterPath     string
	UvPath                    string
	PythonEnvInitTimeout      int
	PythonCompileAllExtraArgs string
	HttpProxy                 string
	HttpsProxy                string
	NoProxy                   string
	PipMirrorUrl              string
	PipPreferBinary           bool
	PipVerbose                bool
	PipExtraArgs              string
	StdoutBufferSize          int
	StdoutMaxBufferSize       int
}

func NewLocalPluginRuntime(config LocalPluginRuntimeConfig) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		defaultPythonInterpreterPath: config.PythonInterpreterPath,
		uvPath:                       config.UvPath,
		pythonEnvInitTimeout:         config.PythonEnvInitTimeout,
		pythonCompileAllExtraArgs:    config.PythonCompileAllExtraArgs,
		HttpProxy:                    config.HttpProxy,
		HttpsProxy:                   config.HttpsProxy,
		NoProxy:                      config.NoProxy,
		pipMirrorUrl:                 config.PipMirrorUrl,
		pipPreferBinary:              config.PipPreferBinary,
		pipVerbose:                   config.PipVerbose,
		pipExtraArgs:                 config.PipExtraArgs,
		stdoutBufferSize:             config.StdoutBufferSize,
		stdoutMaxBufferSize:          config.StdoutMaxBufferSize,
	}
}
