package local_runtime

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/mapping"
)

const (
	ScheduleStatusStopped int32 = iota
	ScheduleStatusRunning
)

type LocalPluginRuntime struct {
	basic_runtime.BasicChecksum
	plugin_entities.PluginRuntime

	// to create a new python virtual environment, we need a default python interpreter
	// by using its venv module
	defaultPythonInterpreterPath string
	uvPath                       string

	appConfig *app.Config

	instanceNums int32 // equivalent to K8s replicas

	sessionToInstanceMap mapping.Map[string, *PluginInstance]

	// always keep the nums of instances equal to instanceNums
	instances []*PluginInstance

	// instanceLocker
	instanceLocker *sync.RWMutex

	// round robin index
	// NOTE: use atomic.AddInt64 and atomic.LoadInt64 to update and read it
	roundRobinIndex int64

	// schedule status
	scheduleStatus int32

	// notifier
	notifiers    []PluginRuntimeNotifier
	notifierLock *sync.Mutex
}

type LocalPluginRuntimeConfig struct {
	PythonInterpreterPath string
	UvPath                string
}

// Type returns the runtime type of the plugin
func (r *LocalPluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
}

// AddNotifier adds a notifier to the runtime
func (r *LocalPluginRuntime) AddNotifier(notifier PluginRuntimeNotifier) {
	r.notifierLock.Lock()
	defer r.notifierLock.Unlock()
	r.notifiers = append(r.notifiers, notifier)
}

// WalkNotifiers walks through all notifiers and calls the corresponding method
func (r *LocalPluginRuntime) WalkNotifiers(callback func(notifier PluginRuntimeNotifier)) {
	r.notifierLock.Lock()
	defer r.notifierLock.Unlock()
	for _, notifier := range r.notifiers {
		callback(notifier)
	}
}
