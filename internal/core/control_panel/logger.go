package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type StandardLogger struct{}

func (l *StandardLogger) OnLocalRuntimeStarting(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) {
	log.Info("local runtime starting: %s", pluginUniqueIdentifier)
}

func (l *StandardLogger) OnLocalRuntimeReady(runtime *local_runtime.LocalPluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("local runtime ready: %s", identity)
}

func (l *StandardLogger) OnLocalRuntimeStartFailed(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier, err error) {
	log.Error("local runtime start failed: %s, error: %s", pluginUniqueIdentifier, err)
}

func (l *StandardLogger) OnLocalRuntimeStop(runtime *local_runtime.LocalPluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("local runtime stop: %s", identity)
}

func (l *StandardLogger) OnLocalRuntimeStopped(runtime *local_runtime.LocalPluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("local runtime stopped: %s", identity)
}

func (l *StandardLogger) OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("debugging runtime connected: %s", identity)
}

func (l *StandardLogger) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("debugging runtime disconnected: %s", identity)
}
