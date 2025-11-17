package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func (c *ControlPanel) setupDebuggingServer(config *app.Config) {
	// construct a debugging server for plugin debugging
	if c.debuggingServer != nil {
		return
	}
	c.debuggingServer = debugging_runtime.NewDebuggingPluginServer(config, c.mediaBucket)

	// setup notifiers
	c.debuggingServer.AddNotifier(&DebuggingRuntimeSignal{
		onConnected:    c.onDebuggingRuntimeConnected,
		onDisconnected: c.onDebuggingRuntimeDisconnected,
	})
}

func (c *ControlPanel) onDebuggingRuntimeConnected(
	rpr *debugging_runtime.RemotePluginRuntime,
) error {
	// handle plugin connection
	pluginIdentifier, err := rpr.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid: %s", err)
		return err
	}

	// store plugin runtime
	c.debuggingPluginRuntime.Store(pluginIdentifier, rpr)

	// notify notifiers a new debugging runtime is connected
	c.WalkNotifiers(func(notifier ControlPanelNotifier) {
		notifier.OnDebuggingRuntimeConnected(rpr)
	})

	return nil
}

func (c *ControlPanel) onDebuggingRuntimeDisconnected(
	rpr *debugging_runtime.RemotePluginRuntime,
) {
	// handle plugin disconnecting
	pluginIdentifier, err := rpr.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid: %s", err)
		return
	}

	// delete plugin runtime
	c.debuggingPluginRuntime.Delete(pluginIdentifier)

	// notify notifiers a new debugging runtime is disconnected
	c.WalkNotifiers(func(notifier ControlPanelNotifier) {
		notifier.OnDebuggingRuntimeDisconnected(rpr)
	})
}

func (c *ControlPanel) startDebuggingServer() error {
	// launch debugging server
	return c.debuggingServer.Launch()
}
