package controlpanel

import (
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/lock"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type ControlPanel struct {
	// app config
	config *app.Config

	// debugging server
	debuggingServer *debugging_runtime.RemotePluginServer

	// media bucket
	mediaBucket *media_transport.MediaBucket

	// installed bucket
	installedBucket *media_transport.InstalledBucket

	// notifiers
	controlPanelNotifiers    []ControlPanelNotifier
	controlPanelNotifierLock *sync.RWMutex

	// local plugin runtimes map
	// plugin unique identifier -> local plugin runtime
	localPluginRuntimes mapping.Map[
		plugin_entities.PluginUniqueIdentifier,
		*local_runtime.LocalPluginRuntime,
	]

	// local plugin launching semaphore
	// we allow multiple plugins to be installed concurrently
	// to control the concurrency, this semaphore is introduced
	localPluginLaunchingSemaphore chan bool

	// how many times a local plugin failed to launch
	// controls retries and waiting time after failures
	localPluginFailsRecord mapping.Map[
		plugin_entities.PluginUniqueIdentifier,
		LocalPluginFailsRecord,
	]

	// local plugin installation lock
	// locks when a plugin is on its installation process, avoid the same plugin
	// to be processed concurrently
	localPluginInstallationLock *lock.GranularityLock

	// debugging plugin runtime
	debuggingPluginRuntime mapping.Map[
		plugin_entities.PluginUniqueIdentifier,
		*debugging_runtime.RemotePluginRuntime,
	]
}

type LocalPluginFailsRecord struct {
	RetryCount  int32
	LastTriedAt time.Time
}

// create a new control panel as the engine of the local plugin daemon
func NewControlPanel(
	config *app.Config,
	mediaBucket *media_transport.MediaBucket,
	installedBucket *media_transport.InstalledBucket,
) *ControlPanel {
	return &ControlPanel{
		config:          config,
		mediaBucket:     mediaBucket,
		installedBucket: installedBucket,

		localPluginLaunchingSemaphore: make(chan bool, config.PluginLocalLaunchingConcurrent),

		// notifiers initialization
		controlPanelNotifiers:    []ControlPanelNotifier{},
		controlPanelNotifierLock: &sync.RWMutex{},

		// local plugin installation lock
		localPluginInstallationLock: lock.NewGranularityLock(),
	}
}
