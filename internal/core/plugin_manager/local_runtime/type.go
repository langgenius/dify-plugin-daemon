package local_runtime

import (
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type launchStage int

const (
	// launchStageInit represents the initial state of a plugin before it has been verified to work properly.
	// launchStageVerified represents the state after a plugin has been successfully started and verified.
	//
	// These states help determine how to handle plugin failures:
	// - When a plugin is in launchStageInit and fails, we return an error from StartPlugin
	//   because we haven't verified it can work properly yet.
	// - When a plugin is in launchStageVerified and fails, we treat it as an unexpected exit
	//   and trigger the automatic restart logic instead of returning an error.
	//
	// This distinction is necessary because the outer layers need to know if plugin startup was successful,
	// while also supporting autoScale and automatic restart features of the plugin runtime.
	launchStageInit launchStage = iota
	launchStageVerifiedWorking
)

type stdioHolderKey struct {
	instanceId string
	attachedAt time.Time
}

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

	isNotFirstStart bool

	// max instances
	maxInstances int
	minInstances int
	autoScale    bool

	stdioHolders []*pluginInstance

	sessionToStdioHolder map[string]*stdioHolderKey
	stdioHolderLock      *sync.Mutex

	stage   launchStage
	scaling bool
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
	AutoScale                 bool
	MaxInstances              int
	MinInstances              int
}

func NewLocalPluginRuntime(config LocalPluginRuntimeConfig) *LocalPluginRuntime {
	maxInstances := config.MaxInstances
	if maxInstances < 1 {
		maxInstances = 1
	}

	minInstances := config.MinInstances
	if minInstances < 1 {
		minInstances = 1
	}

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
		maxInstances:                 maxInstances,
		minInstances:                 minInstances,
		autoScale:                    config.AutoScale,
		sessionToStdioHolder:         make(map[string]*stdioHolderKey),
		stdioHolders:                 make([]*pluginInstance, 0),
		stdioHolderLock:              &sync.Mutex{},
	}
}
