package service

import (
	"errors"
	"fmt"
	"time"

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
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"gorm.io/gorm"
)

// // invalidate plugin installation cache
// TODO: invalidate plugin installation cache
// pluginInstallationCacheKey := helper.PluginInstallationCacheKey(originalPluginUniqueIdentifier.PluginID(), tenantId)
// _, _ = cache.AutoDelete[models.PluginInstallation](pluginInstallationCacheKey)

type InstallPluginResponse struct {
	AllInstalled bool   `json:"all_installed"`
	TaskID       string `json:"task_id"`
}

type InstallPluginOnDoneHandler func(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	declaration *plugin_entities.PluginDeclaration,
	meta map[string]any,
) error

type InstallPluginOnMessageHandler func(
	message installation_entities.PluginInstallResponse,
)

func updateTaskStatus(
	taskId string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	modifier func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus),
) error {
	return db.WithTransaction(func(tx *gorm.DB) error {
		task, err := db.GetOne[models.InstallTask](
			db.WithTransactionContext(tx),
			db.Equal("id", taskId),
			db.WLock(), // write lock, multiple tasks can't update the same task
		)

		if err == db.ErrDatabaseNotFound {
			return nil
		}

		if err != nil {
			return err
		}

		taskPointer := &task
		var pluginStatus *models.InstallTaskPluginStatus
		for i := range task.Plugins {
			if task.Plugins[i].PluginUniqueIdentifier == pluginUniqueIdentifier {
				pluginStatus = &task.Plugins[i]
				break
			}
		}

		if pluginStatus == nil {
			return nil
		}

		modifier(taskPointer, pluginStatus)

		successes := 0
		for _, plugin := range taskPointer.Plugins {
			if plugin.Status == models.InstallTaskStatusSuccess {
				successes++
			}
		}

		if successes == len(taskPointer.Plugins) {
			// update status
			taskPointer.Status = models.InstallTaskStatusSuccess
			// delete the task after 120 seconds without transaction
			time.AfterFunc(120*time.Second, func() {
				db.Delete(taskPointer)
			})
		}
		return db.Update(taskPointer, tx)
	})
}

// read the install task stream and handling its updates
func transformInstallTaskStream(
	installTaskStream *stream.Stream[installation_entities.PluginInstallResponse],
	runtimeType plugin_entities.PluginRuntimeType,
	config *app.Config,
	tenantId string,
	source string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	meta map[string]any,
	task *models.InstallTask,
	declaration *plugin_entities.PluginDeclaration,
) *stream.Stream[installation_entities.PluginInstallResponse] {
	responseStream := stream.NewStream[installation_entities.PluginInstallResponse](128)

	routine.Submit(map[string]string{
		"module":   "service",
		"function": "transformInstallTaskStream",
	}, func() {
		defer responseStream.Close()
		updateTaskStatus(
			task.ID,
			pluginUniqueIdentifier,
			func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				plugin.Status = models.InstallTaskStatusRunning
				plugin.Message = "Installing"
			},
		)
		if err := installTaskStream.Async(func(pir installation_entities.PluginInstallResponse) {
			if pir.Event == installation_entities.PluginInstallEventError {
				updateTaskStatus(
					task.ID,
					pluginUniqueIdentifier,
					func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
						task.Status = models.InstallTaskStatusFailed
						plugin.Status = models.InstallTaskStatusFailed
						plugin.Message = pir.Data
					},
				)
				return
			}

			if pir.Event == installation_entities.PluginInstallEventDone {
				if config.PluginAllowOrphans {
					if err := curd.EnsureGlobalReference(
						pluginUniqueIdentifier,
						tenantId,
						runtimeType,
						declaration,
						source,
						meta,
					); err != nil {
						updateTaskStatus(
							task.ID,
							pluginUniqueIdentifier,
							func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
								task.Status = models.InstallTaskStatusFailed
								plugin.Status = models.InstallTaskStatusFailed
								plugin.Message = err.Error()
							},
						)
						return
					}
				}
			}

			updateTaskStatus(
				task.ID,
				pluginUniqueIdentifier,
				func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					plugin.Status = models.InstallTaskStatusSuccess
					plugin.Message = "Installed"
					task.CompletedPlugins++

					// check if all plugins are installed
					if task.CompletedPlugins == task.TotalPlugins {
						task.Status = models.InstallTaskStatusSuccess
					}
				},
			)
		}); err != nil {
			updateTaskStatus(
				task.ID,
				pluginUniqueIdentifier,
				func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					task.Status = models.InstallTaskStatusFailed
					plugin.Status = models.InstallTaskStatusFailed
					plugin.Message = err.Error()
				},
			)
			responseStream.WriteError(err)
		}
	})

	return responseStream
}

func InstallPluginRuntimeToTenant(
	config *app.Config,
	tenantId string,
	pluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	metas []map[string]any,
) (*InstallPluginResponse, error) {
	response := &InstallPluginResponse{}
	pluginsWaitForInstallation := []plugin_entities.PluginUniqueIdentifier{}
	runtimeType := config.Platform.ToPluginRuntimeType()

	task := &models.InstallTask{
		Status:           models.InstallTaskStatusRunning,
		TenantID:         tenantId,
		TotalPlugins:     len(pluginUniqueIdentifiers),
		CompletedPlugins: 0,
		Plugins:          []models.InstallTaskPluginStatus{},
	}

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

	// submit async tasks
	routine.WithMaxRoutine(5, tasks)

	return response, nil
}

func installOnePluginRuntimeToTenant(
	config *app.Config,
	task *models.InstallTask,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
	runtimeType plugin_entities.PluginRuntimeType,
	manager *plugin_manager.PluginManager,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	declaration, err := helper.CombinedGetPluginDeclaration(
		pluginUniqueIdentifier,
		runtimeType,
	)
	if err != nil {
		return nil, err
	}

	installTaskStream, err := manager.Install(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}

	installTaskStream = transformInstallTaskStream(
		installTaskStream,
		runtimeType,
		config,
		task.TenantID,
		source,
		pluginUniqueIdentifier,
		meta,
		task,
		declaration,
	)

	return installTaskStream, nil
}

func InstallPluginFromIdentifiers(
	config *app.Config,
	tenantId string,
	pluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	metas []map[string]any,
) *entities.Response {
	response, err := InstallPluginRuntimeToTenant(
		config,
		tenantId,
		pluginUniqueIdentifiers,
		source,
		metas,
	)
	if err != nil {
		if errors.Is(err, curd.ErrPluginAlreadyInstalled) {
			return exception.BadRequestError(err).ToResponse()
		}
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(response)
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

/*
 * Decode a plugin from a given identifier, no tenant_id is needed
 * When upload local plugin inside Dify, the second step need to ensure that the plugin is valid
 * So we need to provide a way to decode the plugin and verify the signature
 */
func DecodePluginFromIdentifier(
	config *app.Config,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	// get plugin package and decode again
	manager := plugin_manager.Manager()
	pkgFile, err := manager.GetPackage(pluginUniqueIdentifier)
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}

	zipDecoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(
		pkgFile,
		&decoder.ThirdPartySignatureVerificationConfig{
			Enabled:        config.ThirdPartySignatureVerificationEnabled,
			PublicKeyPaths: config.ThirdPartySignatureVerificationPublicKeys,
		},
	)
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}

	verification, _ := zipDecoder.Verification()
	if verification == nil && zipDecoder.Verified() {
		verification = decoder.DefaultVerification()
	}

	declaration, err := zipDecoder.Manifest()
	if err != nil {
		return exception.BadRequestError(err).ToResponse()
	}

	return entities.NewSuccessResponse(map[string]any{
		"unique_identifier": pluginUniqueIdentifier,
		"manifest":          declaration,
		"verification":      verification,
	})
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

	// install the new plugin runtime
	response, err := InstallPluginRuntimeToTenant(
		config,
		tenantId,
		[]plugin_entities.PluginUniqueIdentifier{newPluginUniqueIdentifier},
		source,
		[]map[string]any{meta},
		func(
			pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
			declaration *plugin_entities.PluginDeclaration,
			meta map[string]any,
		) error {
			originalDeclaration, err := helper.CombinedGetPluginDeclaration(
				originalPluginUniqueIdentifier,
				plugin_entities.PluginRuntimeType(installation.RuntimeType),
			)
			if err != nil {
				return err
			}

			newDeclaration, err := helper.CombinedGetPluginDeclaration(
				newPluginUniqueIdentifier,
				plugin_entities.PluginRuntimeType(installation.RuntimeType),
			)
			if err != nil {
				return err
			}
			// uninstall the original plugin
			upgradeResponse, err := curd.UpgradePlugin(
				tenantId,
				originalPluginUniqueIdentifier,
				newPluginUniqueIdentifier,
				originalDeclaration,
				newDeclaration,
				plugin_entities.PluginRuntimeType(installation.RuntimeType),
				source,
				meta,
			)

			if err != nil {
				return err
			}

			// invalidate plugin installation cache
			pluginInstallationCacheKey := helper.PluginInstallationCacheKey(originalPluginUniqueIdentifier.PluginID(), tenantId)
			_, _ = cache.AutoDelete[models.PluginInstallation](pluginInstallationCacheKey)

			if upgradeResponse.IsOriginalPluginDeleted {
				// delete the plugin if no installation left
				manager := plugin_manager.Manager()
				if string(upgradeResponse.DeletedPlugin.InstallType) == string(
					plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
				) {
					err = manager.UninstallFromLocal(
						plugin_entities.PluginUniqueIdentifier(upgradeResponse.DeletedPlugin.PluginUniqueIdentifier),
					)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(response)
}

func FetchPluginInstallationTasks(
	tenant_id string,
	page int,
	page_size int,
) *entities.Response {
	tasks, err := db.GetAll[models.InstallTask](
		db.Equal("tenant_id", tenant_id),
		db.OrderBy("created_at", true),
		db.Page(page, page_size),
	)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(tasks)
}

func FetchPluginInstallationTask(
	tenant_id string,
	task_id string,
) *entities.Response {
	task, err := db.GetOne[models.InstallTask](
		db.Equal("id", task_id),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(task)
}

func DeletePluginInstallationTask(
	tenant_id string,
	task_id string,
) *entities.Response {
	err := db.DeleteByCondition(
		models.InstallTask{
			Model: models.Model{
				ID: task_id,
			},
			TenantID: tenant_id,
		},
	)

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(true)
}

func DeleteAllPluginInstallationTasks(
	tenant_id string,
) *entities.Response {
	err := db.DeleteByCondition(
		models.InstallTask{
			TenantID: tenant_id,
		},
	)
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(true)
}

func DeletePluginInstallationItemFromTask(
	tenant_id string,
	task_id string,
	identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	err := db.WithTransaction(func(tx *gorm.DB) error {
		item, err := db.GetOne[models.InstallTask](
			db.WithTransactionContext(tx),
			db.Equal("id", task_id),
			db.Equal("tenant_id", tenant_id),
			db.WLock(),
		)

		if err != nil {
			return err
		}

		plugins := []models.InstallTaskPluginStatus{}
		for _, plugin := range item.Plugins {
			if plugin.PluginUniqueIdentifier != identifier {
				plugins = append(plugins, plugin)
			}
		}

		successes := 0
		for _, plugin := range plugins {
			if plugin.Status == models.InstallTaskStatusSuccess {
				successes++
			}
		}

		if len(plugins) == successes {
			// delete the task if all plugins are installed successfully
			err = db.Delete(&item, tx)
		} else {
			item.Plugins = plugins
			err = db.Update(&item, tx)
		}

		return err
	})

	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(true)
}

func FetchPluginFromIdentifier(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	_, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewSuccessResponse(false)
	}
	if err != nil {
		return exception.InternalServerError(err).ToResponse()
	}

	return entities.NewSuccessResponse(true)
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
