package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// // invalidate plugin installation cache
// TODO: invalidate plugin installation cache
// pluginInstallationCacheKey := helper.PluginInstallationCacheKey(originalPluginUniqueIdentifier.PluginID(), tenantId)
// _, _ = cache.AutoDelete[models.PluginInstallation](pluginInstallationCacheKey)

type InstallPluginResponse struct {
	AllInstalled bool   `json:"all_installed"`
	TaskID       string `json:"task_id"`
}

func InstallMultiplePluginsToTenant(
	config *app.Config,
	tenantId string,
	pluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	metas []map[string]any,
) *InstallPluginResponse {

	// TODO: implement this, create task firstly (including EE)
	for i, pluginUniqueIdentifier := range pluginUniqueIdentifiers {
		// fetch plugin declaration first, before installing, we need to ensure pkg is uploaded
		pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			runtimeType,
		)
		if err != nil {
			return nil, err
		}

		// check if plugin is already installed
		_, err = db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
		)

		task.Plugins = append(task.Plugins, models.InstallTaskPluginStatus{
			PluginUniqueIdentifier: pluginUniqueIdentifier,
			PluginID:               pluginUniqueIdentifier.PluginID(),
			Status:                 models.InstallTaskStatusPending,
			Icon:                   pluginDeclaration.Icon,
			IconDark:               pluginDeclaration.IconDark,
			Labels:                 pluginDeclaration.Label,
			Message:                "",
		})

		// found the global plugin installation, no need to start install task
		// just to create a reference to the plugin will make sene
		if err == nil {
			// EE: enterprise only feature, allow orphans
			if config.PluginAllowOrphans {
				if err := curd.EnsureGlobalReference(
					pluginUniqueIdentifier,
					tenantId,
					runtimeType,
					pluginDeclaration,
					source,
					metas[i],
				); err != nil {
					return nil, err
				}
			}

			// just create the reference
			_, _, err = curd.InstallPlugin(
				tenantId,
				pluginUniqueIdentifier,
				runtimeType,
				pluginDeclaration,
				source,
				metas[i],
			)
			if err != nil {
				return nil, errors.Join(err, errors.New("failed on plugin installation"))
			} else {
				task.CompletedPlugins++
				task.Plugins[i].Status = models.InstallTaskStatusSuccess
				task.Plugins[i].Message = "Installed"
			}

			continue
		}

		if err != db.ErrDatabaseNotFound {
			return nil, err
		}

		pluginsWaitForInstallation = append(pluginsWaitForInstallation, pluginUniqueIdentifier)
	}

	if len(pluginsWaitForInstallation) == 0 {
		response.AllInstalled = true
		response.TaskID = ""
		return response, nil
	}

	err := db.Create(task)
	if err != nil {
		return nil, err
	}

	response.TaskID = task.ID
	manager := plugin_manager.Manager()

	tasks := []func(){}
	for i, pluginUniqueIdentifier := range pluginsWaitForInstallation {
		i := i
		tasks = append(tasks, func() {
			installOnePluginRuntimeToTenant(
				config,
				task,
				pluginUniqueIdentifier,
				source,
				metas[i],
				runtimeType,
				manager,
			)
		})
	}

	// TODO: submit async tasks

	return nil, nil
}

func InstallPluginRuntimeToTenant(
	config *app.Config,
	task *models.InstallTask,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
) (*InstallPluginResponse, error) {
	response := &InstallPluginResponse{}
	pluginsWaitForInstallation := []plugin_entities.PluginUniqueIdentifier{}
	runtimeType := config.Platform.ToPluginRuntimeType()

	return response, nil
}

/*
 * Reinstall a plugin from a given identifier, no tenant_id is needed
 */
func ReinstallPluginFromIdentifier(
	ctx *gin.Context,
	config *app.Config,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) {
	baseSSEService(func() (*stream.Stream[installation_entities.PluginInstallResponse], error) {
		pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to get plugin declaration"))
		}

		plugin, err := db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
		)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to get plugin"))
		}

		retStream := stream.NewStream[installation_entities.PluginInstallResponse](128)
		task := &models.InstallTask{
			Status:           models.InstallTaskStatusRunning,
			TenantID:         constants.GlobalTenantId,
			TotalPlugins:     1,
			CompletedPlugins: 0,
			Plugins:          []models.InstallTaskPluginStatus{},
		}
		task.Plugins = append(task.Plugins, models.InstallTaskPluginStatus{
			PluginUniqueIdentifier: pluginUniqueIdentifier,
			PluginID:               pluginUniqueIdentifier.PluginID(),
			Status:                 models.InstallTaskStatusPending,
			Icon:                   pluginDeclaration.Icon,
			IconDark:               pluginDeclaration.IconDark,
			Labels:                 pluginDeclaration.Label,
			Message:                "",
		})

		err = db.Create(task)
		if err != nil {
			return nil, err
		}

		f := func() {
			doInstallPluginRuntime(
				plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
				plugin_manager.Manager(),
				config,
				constants.GlobalTenantId,
				plugin.Source,
				pluginUniqueIdentifier,
				map[string]any{},
				task,
				pluginDeclaration,
				true,
				func(message plugin_manager.PluginInstallResponse) {
					retStream.Write(message)
				},
				func(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier, declaration *plugin_entities.PluginDeclaration, meta map[string]any) error {
					retStream.Close()
					return nil
				})
		}
		routine.Submit(nil, f)
		return retStream, nil
	}, ctx, 1800)
}

func UpgradePlugin(
	config *app.Config,
	tenantId string,
	source string,
	meta map[string]any,
	originalPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	newPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	if originalPluginUniqueIdentifier == newPluginUniqueIdentifier {
		return exception.BadRequestError(errors.New("original and new plugin unique identifier are the same")).ToResponse()
	}

	if originalPluginUniqueIdentifier.PluginID() != newPluginUniqueIdentifier.PluginID() {
		return exception.BadRequestError(errors.New("original and new plugin id are different")).ToResponse()
	}

	// uninstall the original plugin
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenantId),
		db.Equal("plugin_unique_identifier", originalPluginUniqueIdentifier.String()),
		db.Equal("source", source),
	)

	if err == db.ErrDatabaseNotFound {
		return exception.NotFoundError(errors.New("plugin installation not found for this tenant")).ToResponse()
	}

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	// TODO: upgrade process
}

func UninstallPlugin(
	tenant_id string,
	plugin_installation_id string,
) *entities.Response {
	// Check if the plugin exists for the tenant
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("id", plugin_installation_id),
	)
	if err == db.ErrDatabaseNotFound {
		return exception.ErrPluginNotFound().ToResponse()
	}
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(installation.PluginUniqueIdentifier)
	if err != nil {
		return exception.UniqueIdentifierError(err).ToResponse()
	}

	// get declaration
	declaration, err := helper.CombinedGetPluginDeclaration(
		pluginUniqueIdentifier,
		plugin_entities.PluginRuntimeType(installation.RuntimeType),
	)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	// Uninstall the plugin
	deleteResponse, err := curd.UninstallPlugin(
		tenant_id,
		pluginUniqueIdentifier,
		installation.ID,
		declaration,
	)
	if err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to uninstall plugin: %s", err.Error())).ToResponse()
	}

	// invalidate plugin installation cache
	pluginInstallationCacheKey := helper.PluginInstallationCacheKey(pluginUniqueIdentifier.PluginID(), tenant_id)
	_, _ = cache.AutoDelete[models.PluginInstallation](pluginInstallationCacheKey)

	if deleteResponse.IsPluginDeleted {
		// delete the plugin if no installation left
		manager := plugin_manager.Manager()
		if deleteResponse.Installation.RuntimeType == string(
			plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		) {
			err = manager.UninstallFromLocal(pluginUniqueIdentifier)
			if err != nil {
				return exception.InternalServerError(fmt.Errorf("failed to uninstall plugin: %s", err.Error())).ToResponse()
			}
		}
	}

	return entities.NewSuccessResponse(true)
}
