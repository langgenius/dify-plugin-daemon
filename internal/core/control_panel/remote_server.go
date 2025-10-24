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
		// FIXME: it may lead a race condition: while consuming the connection, it's already closed
		identity, err := rpr.Identity()
		if err != nil {
			log.Error("get remote plugin identity failed: %s", err.Error())
			return
		}

		rpr.AddNotifier(&DebuggingRuntimeSignal{
			onDisconnected: func(rpr *debugging_runtime.RemotePluginRuntime) {
				// delete the remote plugin runtime from the control panel
				c.debuggingPluginRuntime.Delete(identity)
			},
		})

		// store the remote plugin runtime
		c.debuggingPluginRuntime.Store(identity, rpr)
	})
}
