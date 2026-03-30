package install_service

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
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
		if !errors.Is(err, curd.ErrPluginAlreadyInstalled) {
			log.Error("install debugging plugin failed", "error", err)
			return
		}

		_, err := runtime.Identity()
		if err != nil {
			log.Error("failed to get plugin identity", "error", err)
			return
		}
		decl := runtime.Configuration()
		pluginID := decl.Author + "/" + decl.Name
		existingInstallation, fetchErr := fetchPluginInstallationByPluginID(runtime.TenantId(), pluginID)
		if fetchErr != nil {
			log.Error("failed to fetch existing installation", "error", fetchErr)
			return
		}
		installation = existingInstallation
	}

	// FIXME(Yeuoly): temporary solution for managing plugin installation model in DB
	runtime.SetInstallationId(installation.ID)
}

func (l *InstallListener) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	var (
		pluginIdentifier plugin_entities.PluginUniqueIdentifier
	)

	installationID := runtime.InstallationId()
	if installationID != "" {
		inst, err := db.GetOne[models.PluginInstallation](
			db.Equal("tenant_id", runtime.TenantId()),
			db.Equal("id", installationID),
		)
		if err == nil && inst.PluginUniqueIdentifier != "" {
			pluginIdentifier, _ = plugin_entities.NewPluginUniqueIdentifier(inst.PluginUniqueIdentifier)
		} else if err != nil && !errors.Is(err, db.ErrDatabaseNotFound) {
			log.Warn("failed to fetch installation for debugging runtime disconnect; falling back to runtime identity", "error", err)
		}
	}

	if pluginIdentifier == "" {
		pi, err := runtime.Identity()
		if err != nil {
			log.Error("failed to get plugin identity, check if your declaration is invalid", "error", err)
		} else {
			pluginIdentifier = pi
		}
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

func fetchPluginInstallationByPluginID(tenantId string, pluginID string) (*models.PluginInstallation, error) {
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenantId),
		db.Equal("plugin_id", pluginID),
	)
	if err != nil {
		return nil, err
	}
	return &installation, nil
}
