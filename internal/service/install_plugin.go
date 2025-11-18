package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/tasks"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

type InstallPluginResponse struct {
	AllInstalled bool   `json:"all_installed"`
	TaskID       string `json:"task_id"`
}

// Dify supports install multiple plugins to a tenant at once
// At most
func InstallMultiplePluginsToTenant(
	config *app.Config,
	tenantId string,
	pluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	metas []map[string]any,
) *entities.Response {
	runtimeType := config.Platform.ToPluginRuntimeType()
	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(errors.New("plugin manager is not initialized")).ToResponse()
	}

	// create len(pluginUniqueIdentifiers) jobs, each job is for one plugin
	// and runs in a single goroutine after the task is created
	jobs := make([]tasks.PluginInstallJob, 0, len(pluginUniqueIdentifiers))
	declarations := make([]*plugin_entities.PluginDeclaration, 0, len(pluginUniqueIdentifiers))
	allInstalled := true

	for i, pluginUniqueIdentifier := range pluginUniqueIdentifiers {
		declaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			runtimeType,
		)
		if err != nil {
			return exception.InternalServerError(errors.Join(err, errors.New("failed to get plugin declaration"))).ToResponse()
		}

		_, err = db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
		)

		needsRuntimeInstall := false
		if err == db.ErrDatabaseNotFound {
			needsRuntimeInstall = true
			allInstalled = false
		} else if err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		job := tasks.PluginInstallJob{
			Identifier:          pluginUniqueIdentifier,
			Declaration:         declaration,
			Meta:                metas[i],
			NeedsRuntimeInstall: needsRuntimeInstall,
		}

		jobs = append(jobs, job)
		declarations = append(declarations, declaration)
	}

	// Always, EE edition needs a global reference to display installation progress
	// add it if needed
	tenants := joinGlobalTenantIfNeeded(config, tenantId)

	// all plugins are installed, no need to create tasks
	// just add DB record and return
	if allInstalled {
		for i := range jobs {
			if err := tasks.SaveInstallationForTenantsToDB(
				tenants,
				jobs[i],
				runtimeType,
				source,
			); err != nil {
				return exception.InternalServerError(errors.Join(err, errors.New("failed on plugin installation"))).ToResponse()
			}
		}

		return entities.NewSuccessResponse(&InstallPluginResponse{
			AllInstalled: true,
			TaskID:       "",
		})
	}

	// create tasks for each plugin
	statuses := buildTaskStatuses(pluginUniqueIdentifiers, declarations)
	taskRegistry, err := createInstallTasks(tenants, statuses)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}
	taskIDs := taskRegistry.IDs()

	for _, job := range jobs {
		jobCopy := job
		// start a new goroutine to install the plugin
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "service",
			routinepkg.RoutineLabelKeyMethod: "InstallPlugin",
		}, func() {
			tasks.ProcessInstallJob(
				manager,
				tenants,
				runtimeType,
				source,
				taskIDs,
				jobCopy,
			)
		})
	}

	return entities.NewSuccessResponse(&InstallPluginResponse{
		AllInstalled: false,
		// EE edition reference task should not be the first one
		// here we use `PrimaryID` to present the user-facing task id
		TaskID: taskRegistry.PrimaryID(),
	})
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

		statuses := buildTaskStatuses(
			[]plugin_entities.PluginUniqueIdentifier{pluginUniqueIdentifier},
			[]*plugin_entities.PluginDeclaration{pluginDeclaration},
		)

		taskRegistry, err := createInstallTasks([]string{constants.GlobalTenantId}, statuses)
		if err != nil {
			return nil, err
		}

		manager := plugin_manager.Manager()
		if manager == nil {
			return nil, errors.New("plugin manager is not initialized")
		}

		retStream := stream.NewStream[installation_entities.PluginInstallResponse](128)
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "service",
			routinepkg.RoutineLabelKeyMethod: "ReinstallPlugin",
		}, func() {
			defer retStream.Close()

			reinstallStream, err := manager.Reinstall(pluginUniqueIdentifier)
			if err != nil {
				tasks.SetTaskStatusForOnePlugin(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, err.Error())
				retStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  err.Error(),
				})
				return
			}

			err = reinstallStream.Process(func(resp installation_entities.PluginInstallResponse) {
				retStream.Write(resp)
				switch resp.Event {
				case installation_entities.PluginInstallEventInfo:
					tasks.SetTaskMessageForOnePlugin(taskRegistry.IDs(), pluginUniqueIdentifier, resp.Data)
				case installation_entities.PluginInstallEventError:
					tasks.SetTaskStatusForOnePlugin(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, resp.Data)
				case installation_entities.PluginInstallEventDone:
					_, _, installErr := curd.InstallPlugin(
						constants.GlobalTenantId,
						pluginUniqueIdentifier,
						plugin.InstallType,
						pluginDeclaration,
						plugin.Source,
						map[string]any{},
					)
					if installErr != nil && installErr != curd.ErrPluginAlreadyInstalled {
						tasks.SetTaskStatusForOnePlugin(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, installErr.Error())
						return
					}
					tasks.SetTaskStatusForOnePlugin(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusSuccess, "Reinstalled")
				}
			})

			if err != nil {
				tasks.SetTaskStatusForOnePlugin(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, err.Error())
				retStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  err.Error(),
				})
			}
		})

		return retStream, nil
	}, ctx, 1800)
}

/*
 * Upgrade a plugin between 2 identifiers
 */
func UpgradePlugin(
	config *app.Config,
	tenantId string,
	source string,
	meta map[string]any,
	originalPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	newPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(errors.New("plugin manager is not initialized")).ToResponse()
	}

	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenantId),
		db.Equal("plugin_unique_identifier", originalPluginUniqueIdentifier.String()),
		db.Equal("source", source),
	)
	if err == db.ErrDatabaseNotFound {
		return exception.NotFoundError(errors.New("plugin installation not found for this tenant")).ToResponse()
	} else if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	runtimeType := plugin_entities.PluginRuntimeType(installation.RuntimeType)
	originalDeclaration, err := helper.CombinedGetPluginDeclaration(originalPluginUniqueIdentifier, runtimeType)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}
	newDeclaration, err := helper.CombinedGetPluginDeclaration(newPluginUniqueIdentifier, runtimeType)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	// check if the new plugin is already installed
	_, err = db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", newPluginUniqueIdentifier.String()),
	)
	if err == nil {
		response, err := curd.UpgradePlugin(
			tenantId,
			originalPluginUniqueIdentifier,
			newPluginUniqueIdentifier,
			originalDeclaration,
			newDeclaration,
			runtimeType,
			source,
			meta,
		)
		if err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		// call RemovePluginIfNeeded in a new goroutine
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "service",
			routinepkg.RoutineLabelKeyMethod: "UpgradePlugin.RemovePluginIfNeeded",
		}, func() {
			if err := tasks.RemovePluginIfNeeded(manager, originalPluginUniqueIdentifier, response); err != nil {
				log.Error("failed to remove uninstalled plugin: %v", err)
			}
		})

		return entities.NewSuccessResponse(&InstallPluginResponse{
			AllInstalled: true,
			TaskID:       "",
		})
	} else if err != db.ErrDatabaseNotFound {
		return exception.InternalServerError(err).ToResponse()
	}

	// construct tenant jobs
	tenants := joinGlobalTenantIfNeeded(config, tenantId)

	job := tasks.PluginUpgradeJob{
		NewIdentifier:       newPluginUniqueIdentifier,
		NewDeclaration:      newDeclaration,
		OriginalIdentifier:  originalPluginUniqueIdentifier,
		OriginalDeclaration: originalDeclaration,
		Meta:                meta,
	}

	statuses := buildTaskStatuses(
		[]plugin_entities.PluginUniqueIdentifier{newPluginUniqueIdentifier},
		[]*plugin_entities.PluginDeclaration{newDeclaration},
	)

	taskRegistry, err := createInstallTasks(tenants, statuses)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	taskIDs := taskRegistry.IDs()

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "service",
		routinepkg.RoutineLabelKeyMethod: "UpgradePlugin",
	}, func() {
		tasks.ProcessUpgradeJob(
			manager,
			tenants,
			runtimeType,
			source,
			taskIDs,
			job,
		)
	})

	return entities.NewSuccessResponse(&InstallPluginResponse{
		AllInstalled: false,
		TaskID:       taskRegistry.PrimaryID(),
	})
}

func UninstallPlugin(
	tenant_id string,
	plugin_installation_id string,
) *entities.Response {
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

	declaration, err := helper.CombinedGetPluginDeclaration(
		pluginUniqueIdentifier,
		plugin_entities.PluginRuntimeType(installation.RuntimeType),
	)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	deleteResponse, err := curd.UninstallPlugin(
		tenant_id,
		pluginUniqueIdentifier,
		installation.ID,
		declaration,
	)
	if err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to uninstall plugin: %s", err.Error())).ToResponse()
	}

	pluginInstallationCacheKey := helper.PluginInstallationCacheKey(pluginUniqueIdentifier.PluginID(), tenant_id)
	_, _ = cache.AutoDelete[models.PluginInstallation](pluginInstallationCacheKey)

	if deleteResponse.IsPluginDeleted && deleteResponse.Plugin != nil && deleteResponse.Plugin.InstallType == plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL {
		manager := plugin_manager.Manager()
		if manager == nil {
			return exception.InternalServerError(errors.New("plugin manager is not initialized")).ToResponse()
		}

		if err := manager.RemoveLocalPlugin(pluginUniqueIdentifier); err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		shutdownCh, err := manager.ShutdownLocalPluginGracefully(pluginUniqueIdentifier)
		if err == controlpanel.ErrLocalPluginRuntimeNotFound {
			return entities.NewSuccessResponse(true)
		} else if err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		if err := waitGracefulShutdown(shutdownCh); err != nil {
			return exception.InternalServerError(err).ToResponse()
		}
	}

	return entities.NewSuccessResponse(true)
}

func waitGracefulShutdown(ch <-chan error) error {
	if ch == nil {
		return nil
	}

	for err := range ch {
		if err != nil {
			return err
		}
	}

	return nil
}
