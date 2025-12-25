package controlpanel

import (
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type StandardLogger struct{}

func (l *StandardLogger) OnLocalRuntimeStarting(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) {
	log.Info("local runtime starting", "plugin", pluginUniqueIdentifier)
}

func (l *StandardLogger) OnLocalRuntimeReady(runtime *local_runtime.LocalPluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("local runtime ready", "plugin", identity)
}

func (l *StandardLogger) OnLocalRuntimeStartFailed(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier, err error) {
	log.Error("local runtime start failed", "plugin", pluginUniqueIdentifier, "error", err)
}

func (l *StandardLogger) OnLocalRuntimeStop(runtime *local_runtime.LocalPluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("local runtime stop", "plugin", identity)
}

func (l *StandardLogger) OnLocalRuntimeStopped(runtime *local_runtime.LocalPluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("local runtime stopped", "plugin", identity)
}

var (
	loggers = map[string]func(string, ...any){
		"debug":    log.Debug,
		"warn":     log.Warn,
		"warning":  log.Warn,
		"error":    log.Error,
		"fatal":    log.Error,
		"critical": log.Error,
	}
)

func (l *StandardLogger) OnLocalRuntimeInstanceLog(
	runtime *local_runtime.LocalPluginRuntime,
	instance *local_runtime.PluginInstance,
	event plugin_entities.PluginLogEvent,
) {
	// notify terminal
	instanceID := instance.ID()[:8]
	loggerFunc, ok := loggers[strings.ToLower(event.Level)]
	if !ok {
		loggerFunc = log.Info
	}
	identity, _ := runtime.Identity()

	loggerFunc("plugin instance log", "plugin", identity, "instance", instanceID, "message", event.Message)
}

func (l *StandardLogger) OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("debugging runtime connected", "plugin", identity)
}

func (l *StandardLogger) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("debugging runtime disconnected", "plugin", identity)
}

func (l *StandardLogger) OnLocalRuntimeScaleUp(runtime *local_runtime.LocalPluginRuntime, instanceNums int32) {
	identity, _ := runtime.Identity()
	log.Info("local runtime scale up", "plugin", identity, "instance_nums", instanceNums)
}

func (l *StandardLogger) OnLocalRuntimeScaleDown(runtime *local_runtime.LocalPluginRuntime, instanceNums int32) {
	identity, _ := runtime.Identity()
	log.Info("local runtime scale down", "plugin", identity, "instance_nums", instanceNums)
}
