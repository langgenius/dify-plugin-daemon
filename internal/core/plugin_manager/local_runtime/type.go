package local_runtime

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalPluginRuntime struct {
	basic_runtime.BasicChecksum
	plugin_entities.PluginRuntime

	waitChan chan bool

	// python env init timeout
	pythonEnvInitTimeout int

	// python compileall extra args
	pythonCompileAllExtraArgs string

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	defaultPythonInterpreterPath string
	uvPath                       string

	waitChanLock    sync.Mutex
	waitStartedChan []chan bool
	waitStoppedChan []chan bool

	isNotFirstStart bool

	appConfig *app.Config

	instanceNums int // equivalent to K8s replicas

	// always keep the nums of instances equal to instanceNums
	instances []*PluginInstance
}

type LocalPluginRuntimeConfig struct {
	PythonInterpreterPath     string
	UvPath                    string
	PythonEnvInitTimeout      int
	PythonCompileAllExtraArgs string
}

func NewLocalPluginRuntime(
	config LocalPluginRuntimeConfig,
	appConfig *app.Config,
) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		defaultPythonInterpreterPath: config.PythonInterpreterPath,
		uvPath:                       config.UvPath,
		pythonEnvInitTimeout:         config.PythonEnvInitTimeout,
		pythonCompileAllExtraArgs:    config.PythonCompileAllExtraArgs,
		appConfig:                    appConfig,
	}
}
