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

var (
	loggers = map[string]func(string, ...interface{}){
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

	loggerFunc("plugin %s: instance %s log: %s", identity, instanceID, event.Message)
}

func (l *StandardLogger) OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("debugging runtime connected: %s", identity)
}

func (l *StandardLogger) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	identity, _ := runtime.Identity()
	log.Info("debugging runtime disconnected: %s", identity)
}

func (l *StandardLogger) OnLocalRuntimeScaleUp(runtime *local_runtime.LocalPluginRuntime, instanceNums int32) {
	identity, _ := runtime.Identity()
	log.Info("local runtime scale up: %s, instance nums: %d", identity, instanceNums)
}

func (l *StandardLogger) OnLocalRuntimeScaleDown(runtime *local_runtime.LocalPluginRuntime, instanceNums int32) {
	identity, _ := runtime.Identity()
	log.Info("local runtime scale down: %s, instance nums: %d", identity, instanceNums)
}
