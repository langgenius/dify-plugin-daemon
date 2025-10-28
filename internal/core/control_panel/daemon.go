package controlpanel

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
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

	// local plugin runtimes
	localPluginRuntimes mapping.Map[
		plugin_entities.PluginUniqueIdentifier,
		*local_runtime.LocalPluginRuntime,
	]

	// debugging plugin runtime
	debuggingPluginRuntime mapping.Map[
		plugin_entities.PluginUniqueIdentifier,
		*debugging_runtime.RemotePluginRuntime,
	]
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

		// notifiers initialization
		controlPanelNotifiers:    []ControlPanelNotifier{},
		controlPanelNotifierLock: &sync.RWMutex{},
	}
}
