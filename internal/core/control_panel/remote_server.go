package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func (c *ControlPanel) setupDebuggingServer(config *app.Config) {
	// construct a debugging server for plugin debugging
	if c.debuggingServer != nil {
		return
	}
	c.debuggingServer = debugging_runtime.NewRemotePluginServer(config, c.mediaBucket)

	// setup notifiers
	c.debuggingServer.AddNotifier(&DebuggingRuntimeSignal{
		onConnected:    c.onDebuggingRuntimeConnected,
		onDisconnected: c.OnDebuggingRuntimeDisconnected,
	})
}

func (c *ControlPanel) onDebuggingRuntimeConnected(
	rpr *debugging_runtime.RemotePluginRuntime,
) error {
	// handle plugin connection
	pluginIdentifier, err := rpr.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid: %s", err)
	}
	c.debuggingPluginRuntime.Store(pluginIdentifier, rpr)
}

func (c *ControlPanel) OnDebuggingRuntimeDisconnected(
	rpr *debugging_runtime.RemotePluginRuntime,
) {
	// handle plugin disconnecting
	pluginIdentifier, err := rpr.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid: %s", err)
	}
	c.debuggingPluginRuntime.Delete(pluginIdentifier)
}

func (c *ControlPanel) startDebuggingServer() error {
	// launch debugging server
	return c.debuggingServer.Launch()
}

// consumeDebuggingPlugin consumes TCP connections from clients
// and schedule it to the control panel
func (c *ControlPanel) consumeDebuggingPlugin() {
	// listen to new connections
	c.debuggingServer.Wrap(func(rpr *debugging_runtime.RemotePluginRuntime) {
		identity, err := rpr.Identity()
		if err != nil {
			log.Error("get remote plugin identity failed: %s", err.Error())
			return
		}
	})
}
