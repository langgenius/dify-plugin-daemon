package install_service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type InstallListener struct{}

func (l *InstallListener) OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime) {
	_, installation, err := InstallPlugin(
		runtime.TenantId(),
		"",
		runtime,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE),
		map[string]any{},
	)
	if err != nil {
		log.Error("install debugging plugin failed, error: %v", err)
		return
	}

	// FIXME(Yeuoly): temporary solution for managing plugin installation model in DB
	runtime.SetInstallationId(installation.ID)
}

func (l *InstallListener) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	pluginIdentifier, err := runtime.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid: %s", err)
	}

	if err := UninstallPlugin(
		runtime.TenantId(),
		runtime.InstallationId(),
		pluginIdentifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
	); err != nil {
		log.Error("uninstall debugging plugin failed, error: %v", err)
	}
}

func (l *InstallListener) OnLocalRuntimeReady(runtime *local_runtime.LocalPluginRuntime) {

}

func (l *InstallListener) OnLocalRuntimeStartFailed(runtime *local_runtime.LocalPluginRuntime) {

}
