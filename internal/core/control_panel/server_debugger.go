package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/service/install_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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
	}

	// store plugin runtime
	c.debuggingPluginRuntime.Store(pluginIdentifier, rpr)

	_, installation, err := install_service.InstallPlugin(
		rpr.TenantId(),
		"",
		rpr,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE),
		map[string]any{},
	)
	if err != nil {
		return err
	}

	// FIXME(Yeuoly): temporary solution for managing plugin installation model in DB
	rpr.SetInstallationId(installation.ID)

	return nil
}

func (c *ControlPanel) onDebuggingRuntimeDisconnected(
	rpr *debugging_runtime.RemotePluginRuntime,
) {
	// handle plugin disconnecting
	pluginIdentifier, err := rpr.Identity()
	if err != nil {
		log.Error("failed to get plugin identity, check if your declaration is invalid: %s", err)
	}

	// delete plugin runtime
	c.debuggingPluginRuntime.Delete(pluginIdentifier)

	if err := install_service.UninstallPlugin(
		rpr.TenantId(),
		rpr.InstallationId(),
		pluginIdentifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
	); err != nil {
		log.Error("uninstall debugging plugin failed, error: %v", err)
	}
}

func (c *ControlPanel) startDebuggingServer() error {
	// launch debugging server
	return c.debuggingServer.Launch()
}
