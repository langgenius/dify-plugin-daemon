package local_runtime

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// construct a plugin runtime
// 1. please make sure correct plugin working path is provided
func ConstructPluginRuntime(
	appConfig *app.Config,
	pluginDecoder decoder.PluginDecoder,
) (*LocalPluginRuntime, error) {
	// get manifest
	manifest, err := pluginDecoder.Manifest()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("get plugin manifest error"))
	}

	pluginWorkingPath, err := buildPluginWorkingPath(appConfig, pluginDecoder)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("build plugin working path error"))
	}

	runtime := &LocalPluginRuntime{
		PluginRuntime: plugin_entities.PluginRuntime{
			Config: manifest,
			State: plugin_entities.PluginRuntimeState{
				Status:      plugin_entities.PLUGIN_RUNTIME_STATUS_PENDING,
				Restarts:    0,
				ActiveAt:    nil,
				Verified:    manifest.Verified,
				WorkingPath: pluginWorkingPath,
			},
		},
		BasicChecksum: basic_runtime.BasicChecksum{
			Decoder: pluginDecoder,
		},
		scheduleStatus:               ScheduleStatusStopped,
		defaultPythonInterpreterPath: appConfig.PythonInterpreterPath,
		uvPath:                       appConfig.UvPath,
		appConfig:                    appConfig,

		instances:      []*PluginInstance{},
		instanceLocker: &sync.RWMutex{},

		notifiers:    []PluginRuntimeNotifier{},
		notifierLock: &sync.Mutex{},
	}
	return runtime, nil
}

// generate plugin working path using author/name@checksum, but replace : with -
// some platform like windows may not allow : in the path
func buildPluginWorkingPath(
	appConfig *app.Config,
	pluginDecoder decoder.PluginDecoder,
) (string, error) {
	manifest, err := pluginDecoder.Manifest()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("get plugin manifest error"))
	}

	checksum, err := pluginDecoder.Checksum()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("calculate checksum error"))
	}

	identity := manifest.Identity()
	identity = strings.ReplaceAll(identity, ":", "-")

	return path.Join(
		appConfig.PluginWorkingPath,
		fmt.Sprintf("%s@%s", identity, checksum),
	), nil
}
