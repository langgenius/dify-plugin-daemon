package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
)

type DebuggingRuntimeSignal struct {
	onConnected    func(rpr *debugging_runtime.RemotePluginRuntime)
	onDisconnected func(rpr *debugging_runtime.RemotePluginRuntime)
}

func (c *DebuggingRuntimeSignal) OnRuntimeConnected(rpr *debugging_runtime.RemotePluginRuntime) {
	if c.onConnected != nil {
		c.onConnected(rpr)
	}
}

func (c *DebuggingRuntimeSignal) OnRuntimeDisconnected(rpr *debugging_runtime.RemotePluginRuntime) {
	if c.onDisconnected != nil {
		c.onDisconnected(rpr)
	}
}
