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
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type InstallPluginResponse struct {
	AllInstalled bool   `json:"all_installed"`
	TaskID       string `json:"task_id"`
}

type pluginInstallJob struct {
	identifier          plugin_entities.PluginUniqueIdentifier
	declaration         *plugin_entities.PluginDeclaration
	meta                map[string]any
	needsRuntimeInstall bool
}

type installTaskRegistry struct {
	order []string
	tasks map[string]*models.InstallTask
}

func (r *installTaskRegistry) IDs() []string {
	ids := make([]string, 0, len(r.order))
	for _, tenantID := range r.order {
		if task, ok := r.tasks[tenantID]; ok {
			ids = append(ids, task.ID)
		}
	}
	return ids
}

func (r *installTaskRegistry) PrimaryID() string {
	if len(r.order) == 0 {
		return ""
	}
	if task, ok := r.tasks[r.order[0]]; ok {
		return task.ID
	}
	return ""
}

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

	jobs := make([]pluginInstallJob, 0, len(pluginUniqueIdentifiers))
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

		job := pluginInstallJob{
			identifier:          pluginUniqueIdentifier,
			declaration:         declaration,
			meta:                metas[i],
			needsRuntimeInstall: needsRuntimeInstall,
		}

		jobs = append(jobs, job)
		declarations = append(declarations, declaration)
	}

	tenants := resolveInstallTenants(config, tenantId)

	if allInstalled {
		for i := range jobs {
			if err := installForTenants(
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

	statuses := buildTaskStatuses(pluginUniqueIdentifiers, declarations)
	taskRegistry, err := createInstallTasks(tenants, statuses)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}
	taskIDs := taskRegistry.IDs()

	for _, job := range jobs {
		jobCopy := job
		routine.Submit(map[string]string{
			"module": "service",
			"func":   "InstallPlugin",
		}, func() {
			processInstallJob(
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
		TaskID:       taskRegistry.PrimaryID(),
	})
}

func processInstallJob(
	manager *plugin_manager.PluginManager,
	tenants []string,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
	taskIDs []string,
	job pluginInstallJob,
) {
	if !job.needsRuntimeInstall {
		if err := installForTenants(tenants, job, runtimeType, source); err != nil {
			setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusFailed, err.Error())
			return
		}
		setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusSuccess, "Installed")
		return
	}

	installationStream, err := manager.Install(job.identifier)
	if err != nil {
		setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusFailed, fmt.Sprintf("failed to start installation: %v", err))
		return
	}

	err = installationStream.Async(func(resp installation_entities.PluginInstallResponse) {
		switch resp.Event {
		case installation_entities.PluginInstallEventInfo:
			setTaskMessage(taskIDs, job.identifier, resp.Data)
		case installation_entities.PluginInstallEventError:
			setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusFailed, resp.Data)
		case installation_entities.PluginInstallEventDone:
			if err := installForTenants(tenants, job, runtimeType, source); err != nil {
				setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusFailed, err.Error())
				return
			}
			setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusSuccess, "Installed")
		}
	})
	if err != nil {
		setTaskStatus(taskIDs, job.identifier, models.InstallTaskStatusFailed, err.Error())
	}
}

func installForTenants(
	tenants []string,
	job pluginInstallJob,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
) error {
	for _, tenantID := range tenants {
		if err := installForTenant(tenantID, job, runtimeType, source); err != nil {
			return err
		}
	}
	return nil
}

func installForTenant(
	tenantID string,
	job pluginInstallJob,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
) error {
	_, _, err := curd.InstallPlugin(
		tenantID,
		job.identifier,
		runtimeType,
		job.declaration,
		source,
		job.meta,
	)
	if err != nil && err != curd.ErrPluginAlreadyInstalled {
		return err
	}
	return nil
}

func buildTaskStatuses(
	pluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier,
	declarations []*plugin_entities.PluginDeclaration,
) []models.InstallTaskPluginStatus {
	statuses := make([]models.InstallTaskPluginStatus, len(pluginUniqueIdentifiers))
	for i, identifier := range pluginUniqueIdentifiers {
		statuses[i] = models.InstallTaskPluginStatus{
			PluginUniqueIdentifier: identifier,
			PluginID:               identifier.PluginID(),
			Status:                 models.InstallTaskStatusPending,
			Icon:                   declarations[i].Icon,
			IconDark:               declarations[i].IconDark,
			Labels:                 declarations[i].Label,
			Message:                "",
		}
	}
	return statuses
}

func createInstallTasks(
	tenants []string,
	statuses []models.InstallTaskPluginStatus,
) (*installTaskRegistry, error) {
	registry := &installTaskRegistry{
		order: append([]string{}, tenants...),
		tasks: make(map[string]*models.InstallTask, len(tenants)),
	}

	for _, tenantID := range tenants {
		statusCopy := make([]models.InstallTaskPluginStatus, len(statuses))
		copy(statusCopy, statuses)

		task := &models.InstallTask{
			Status:           models.InstallTaskStatusRunning,
			TenantID:         tenantID,
			TotalPlugins:     len(statusCopy),
			CompletedPlugins: 0,
			Plugins:          statusCopy,
		}

		if err := db.Create(task); err != nil {
			return nil, err
		}

		registry.tasks[tenantID] = task
	}

	return registry, nil
}

func setTaskStatus(
	taskIDs []string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	status models.InstallTaskStatus,
	message string,
) {
	for _, taskID := range taskIDs {
		if err := updateTaskStatus(taskID, pluginUniqueIdentifier, func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
			previousStatus := plugin.Status
			plugin.Status = status
			plugin.Message = message
			if status == models.InstallTaskStatusSuccess && previousStatus != models.InstallTaskStatusSuccess {
				task.CompletedPlugins++
			}
			if status == models.InstallTaskStatusFailed {
				task.Status = models.InstallTaskStatusFailed
				if previousStatus == models.InstallTaskStatusSuccess && task.CompletedPlugins > 0 {
					task.CompletedPlugins--
				}
			}
		}); err != nil {
			log.Error("failed to update task status for %s: %v", pluginUniqueIdentifier.String(), err)
		}
	}
}

func setTaskMessage(
	taskIDs []string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	message string,
) {
	for _, taskID := range taskIDs {
		if err := updateTaskStatus(taskID, pluginUniqueIdentifier, func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
			plugin.Message = message
		}); err != nil {
			log.Error("failed to update task message for %s: %v", pluginUniqueIdentifier.String(), err)
		}
	}
}

func resolveInstallTenants(config *app.Config, tenantId string) []string {
	tenants := []string{tenantId}
	if tenantId != constants.GlobalTenantId && config.PluginAllowOrphans {
		tenants = append(tenants, constants.GlobalTenantId)
	}
	return tenants
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
		routine.Submit(map[string]string{
			"module": "service",
			"func":   "ReinstallPlugin",
		}, func() {
			defer retStream.Close()

			reinstallStream, err := manager.Reinstall(pluginUniqueIdentifier)
			if err != nil {
				setTaskStatus(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, err.Error())
				retStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  err.Error(),
				})
				return
			}

			err = reinstallStream.Async(func(resp installation_entities.PluginInstallResponse) {
				retStream.Write(resp)
				switch resp.Event {
				case installation_entities.PluginInstallEventInfo:
					setTaskMessage(taskRegistry.IDs(), pluginUniqueIdentifier, resp.Data)
				case installation_entities.PluginInstallEventError:
					setTaskStatus(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, resp.Data)
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
						setTaskStatus(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, installErr.Error())
						return
					}
					setTaskStatus(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusSuccess, "Reinstalled")
				}
			})

			if err != nil {
				setTaskStatus(taskRegistry.IDs(), pluginUniqueIdentifier, models.InstallTaskStatusFailed, err.Error())
				retStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  err.Error(),
				})
			}
		})

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

	runtimeType := plugin_entities.PluginRuntimeType(installation.RuntimeType)

	originalDeclaration, err := helper.CombinedGetPluginDeclaration(originalPluginUniqueIdentifier, runtimeType)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	newDeclaration, err := helper.CombinedGetPluginDeclaration(newPluginUniqueIdentifier, runtimeType)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

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

	if response.IsOriginalPluginDeleted && response.DeletedPlugin != nil && response.DeletedPlugin.InstallType == plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL {
		manager := plugin_manager.Manager()
		if manager == nil {
			return exception.InternalServerError(errors.New("plugin manager is not initialized")).ToResponse()
		}

		if err := manager.RemoveLocalPlugin(originalPluginUniqueIdentifier); err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		shutdownCh, err := manager.ShutdownLocalPluginGracefully(originalPluginUniqueIdentifier)
		if err != nil {
			return exception.InternalServerError(err).ToResponse()
		}

		if err := waitGracefulShutdown(shutdownCh); err != nil {
			return exception.InternalServerError(err).ToResponse()
		}
	}

	return entities.NewSuccessResponse(true)
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
		if err != nil {
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
