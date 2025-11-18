package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
)

type DebuggingRuntimeSignal struct {
	// Triggers if a new client connection established
	onConnected func(rpr *debugging_runtime.RemotePluginRuntime) error

	// Triggers if connection lost
	onDisconnected func(rpr *debugging_runtime.RemotePluginRuntime)

	// Triggers if the server is shutting down
	onServerShutdown func(reason debugging_runtime.ServerShutdownReason)
}

func (c *DebuggingRuntimeSignal) OnRuntimeConnected(rpr *debugging_runtime.RemotePluginRuntime) error {
	if c.onConnected != nil {
		return c.onConnected(rpr)
	}
	return nil
}

func (c *DebuggingRuntimeSignal) OnRuntimeDisconnected(rpr *debugging_runtime.RemotePluginRuntime) {
	if c.onDisconnected != nil {
		c.onDisconnected(rpr)
	}
}

func (c *DebuggingRuntimeSignal) OnServerShutdown(reason debugging_runtime.ServerShutdownReason) {
	if c.onServerShutdown != nil {
		c.onServerShutdown(reason)
	}
}
