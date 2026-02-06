package install_service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type InstallListener struct{}

func (l *InstallListener) OnLocalRuntimeStarting(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) {
}

func (l *InstallListener) OnLocalRuntimeReady(runtime *local_runtime.LocalPluginRuntime) {
}

func (l *InstallListener) OnLocalRuntimeStartFailed(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	err error,
) {
}

func (l *InstallListener) OnLocalRuntimeStop(runtime *local_runtime.LocalPluginRuntime) {
}

func (l *InstallListener) OnLocalRuntimeStopped(runtime *local_runtime.LocalPluginRuntime) {
}

func (l *InstallListener) OnLocalRuntimeScaleUp(runtime *local_runtime.LocalPluginRuntime, instanceNums int32) {
}

func (l *InstallListener) OnLocalRuntimeScaleDown(runtime *local_runtime.LocalPluginRuntime, instanceNums int32) {
}

func (l *InstallListener) OnLocalRuntimeInstanceLog(
	runtime *local_runtime.LocalPluginRuntime,
	instance *local_runtime.PluginInstance,
	event plugin_entities.PluginLogEvent,
) {
}

func (l *InstallListener) OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime) {
	_, installation, err := InstallPlugin(
		runtime.TenantId(),
		"",
		runtime,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE),
		map[string]any{},
	)
	if err != nil {
		log.Error("install debugging plugin failed", "error", err)
		return
	}

	// FIXME(Yeuoly): temporary solution for managing plugin installation model in DB
	runtime.SetInstallationId(installation.ID)
}

func (l *InstallListener) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	pluginIdentifier, err := runtime.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid", "error", err)
	}

	if err := UninstallPlugin(
		runtime.TenantId(),
		runtime.InstallationId(),
		pluginIdentifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
	); err != nil {
		log.Error("uninstall debugging plugin failed", "error", err)
	}
}
