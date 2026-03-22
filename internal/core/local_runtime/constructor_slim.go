package local_runtime

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// NewLocalPluginRuntime creates a LocalPluginRuntime with a known working path.
// Unlike ConstructPluginRuntime, it does not compute a checksum-based path,
// which avoids walking the full directory tree (including .venv / .uv-cache).
func NewLocalPluginRuntime(
	appConfig *app.Config,
	pluginDecoder decoder.PluginDecoder,
	manifest plugin_entities.PluginDeclaration,
	workingPath string,
) *LocalPluginRuntime {
	return &LocalPluginRuntime{
		PluginRuntime: plugin_entities.PluginRuntime{
			Config: manifest,
			State: plugin_entities.PluginRuntimeState{
				Status:      plugin_entities.PLUGIN_RUNTIME_STATUS_PENDING,
				Verified:    manifest.Verified,
				WorkingPath: workingPath,
			},
		},
		BasicChecksum: basic_runtime.BasicChecksum{
			Decoder: pluginDecoder,
		},
		scheduleStatus:               ScheduleStatusStopped,
		defaultPythonInterpreterPath: appConfig.PythonInterpreterPath,
		uvPath:                       appConfig.UvPath,
		appConfig:                    appConfig,
		instances:                    []*PluginInstance{},
		instanceLocker:               &sync.RWMutex{},
		notifiers:                    []PluginRuntimeNotifier{},
		notifierLock:                 &sync.Mutex{},
	}
}
