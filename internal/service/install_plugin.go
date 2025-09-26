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
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"gorm.io/gorm"
)

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
	message plugin_manager.PluginInstallResponse,
)

func doInstallPluginRuntime(
	runtimeType plugin_entities.PluginRuntimeType,
	manager *plugin_manager.PluginManager,
	config *app.Config,
	tenant_id string,
	source string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	meta map[string]any,
	task *models.InstallTask,
	declaration *plugin_entities.PluginDeclaration,
	reinstall bool,
	onMessage InstallPluginOnMessageHandler,
	onDone InstallPluginOnDoneHandler,
) {
	var err error
	updateTaskStatus := func(modifier func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus)) {

		if err := db.WithTransaction(func(tx *gorm.DB) error {
			task, err := db.GetOne[models.InstallTask](
				db.WithTransactionContext(tx),
				db.Equal("id", task.ID),
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
		}); err != nil {
			log.Error("failed to update install task status %s", err.Error())
		}
	}

	updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
		plugin.Status = models.InstallTaskStatusRunning
		plugin.Message = "Installing"
	})

	var stream *stream.Stream[plugin_manager.PluginInstallResponse]
	if config.Platform == app.PLATFORM_SERVERLESS {
		var zipDecoder *decoder.ZipPluginDecoder
		var pkgFile []byte

		pkgFile, err = manager.GetPackage(pluginUniqueIdentifier)
		if err != nil {
			updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				task.Status = models.InstallTaskStatusFailed
				plugin.Status = models.InstallTaskStatusFailed
				plugin.Message = "Failed to read plugin package"
				onMessage(plugin_manager.PluginInstallResponse{
					Event: plugin_manager.PluginInstallEventError,
					Data:  plugin.Message,
				})
			})
			return
		}

		zipDecoder, err = decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(
			pkgFile,
			&decoder.ThirdPartySignatureVerificationConfig{
				Enabled:        config.ThirdPartySignatureVerificationEnabled,
				PublicKeyPaths: config.ThirdPartySignatureVerificationPublicKeys,
			},
		)
		if err != nil {
			updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				task.Status = models.InstallTaskStatusFailed
				plugin.Status = models.InstallTaskStatusFailed
				plugin.Message = err.Error()
				onMessage(plugin_manager.PluginInstallResponse{
					Event: plugin_manager.PluginInstallEventError,
					Data:  plugin.Message,
				})
			})
			return
		}
		if reinstall {
			stream, err = manager.ReinstallToServerlessFromPkg(pkgFile, zipDecoder)
		} else {
			stream, err = manager.InstallToServerlessFromPkg(pkgFile, zipDecoder, source, meta)
		}
	} else if config.Platform == app.PLATFORM_LOCAL {
		if reinstall {
			log.Warn("reinstall is not supported on local platform, will do install")
		}
		stream, err = manager.InstallToLocal(pluginUniqueIdentifier, source, meta)
	} else {
		updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
			task.Status = models.InstallTaskStatusFailed
			plugin.Status = models.InstallTaskStatusFailed
			plugin.Message = "Unsupported platform"
			onMessage(plugin_manager.PluginInstallResponse{
				Event: plugin_manager.PluginInstallEventError,
				Data:  plugin.Message,
			})
		})
		return
	}

	if err != nil {
		updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
			task.Status = models.InstallTaskStatusFailed
			plugin.Status = models.InstallTaskStatusFailed
			plugin.Message = err.Error()
			onMessage(plugin_manager.PluginInstallResponse{
				Event: plugin_manager.PluginInstallEventError,
				Data:  plugin.Message,
			})
		})
		return
	}

	for stream.Next() {
		message, err := stream.Read()
		if err != nil {
			updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				task.Status = models.InstallTaskStatusFailed
				plugin.Status = models.InstallTaskStatusFailed
				plugin.Message = err.Error()
			})
			return
		}
		onMessage(message)
		if message.Event == plugin_manager.PluginInstallEventError {
			updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				task.Status = models.InstallTaskStatusFailed
				plugin.Status = models.InstallTaskStatusFailed
				plugin.Message = message.Data
			})
			return
		}

		if message.Event == plugin_manager.PluginInstallEventDone {
			if err := curd.EnsureGlobalReferenceIfRequired(pluginUniqueIdentifier, tenant_id, runtimeType, declaration, source, meta); err != nil {
				updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					task.Status = models.InstallTaskStatusFailed
					plugin.Status = models.InstallTaskStatusFailed
					plugin.Message = err.Error()
				})
				return
			}
			if err := onDone(pluginUniqueIdentifier, declaration, meta); err != nil {
				updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					task.Status = models.InstallTaskStatusFailed
					plugin.Status = models.InstallTaskStatusFailed
					plugin.Message = "Failed to create plugin, perhaps it's already installed"
				})
				return
			}
		}
	}

	updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
		plugin.Status = models.InstallTaskStatusSuccess
		plugin.Message = "Installed"
		task.CompletedPlugins++

		// check if all plugins are installed
		if task.CompletedPlugins == task.TotalPlugins {
			task.Status = models.InstallTaskStatusSuccess
		}
	})
}

func InstallPluginRuntimeToTenant(
	config *app.Config,
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	metas []map[string]any,
	onDone InstallPluginOnDoneHandler, // since installing plugin is a async task, we need to call it asynchronously
) (*InstallPluginResponse, error) {
	response := &InstallPluginResponse{}
	pluginsWaitForInstallation := []plugin_entities.PluginUniqueIdentifier{}

	runtimeType := plugin_entities.PluginRuntimeType("")
	if config.Platform == app.PLATFORM_SERVERLESS {
		runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS
	} else if config.Platform == app.PLATFORM_LOCAL {
		runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
	} else {
		return nil, fmt.Errorf("unsupported platform: %s", config.Platform)
	}

	task := &models.InstallTask{
		Status:           models.InstallTaskStatusRunning,
		TenantID:         tenant_id,
		TotalPlugins:     len(plugin_unique_identifiers),
		CompletedPlugins: 0,
		Plugins:          []models.InstallTaskPluginStatus{},
	}

	for i, pluginUniqueIdentifier := range plugin_unique_identifiers {
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

		if err == nil {
			if err := curd.EnsureGlobalReferenceIfRequired(pluginUniqueIdentifier, tenant_id, runtimeType, pluginDeclaration, source, metas[i]); err != nil {
				return nil, err
			}
			if err := onDone(pluginUniqueIdentifier, pluginDeclaration, metas[i]); err != nil {
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
		declaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			runtimeType,
		)
		if err != nil {
			return nil, err
		}

		i := i
		tasks = append(tasks, func() {
			doInstallPluginRuntime(
				runtimeType,
				manager,
				config,
				tenant_id,
				source,
				pluginUniqueIdentifier,
				metas[i],
				task,
				declaration,
				false,
				func(message plugin_manager.PluginInstallResponse) {},
				onDone)
		})
	}

	// submit async tasks
	routine.WithMaxRoutine(5, tasks)

	return response, nil
}

func InstallPluginFromIdentifiers(
	config *app.Config,
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	metas []map[string]any,
) *entities.Response {
	response, err := InstallPluginRuntimeToTenant(
		config,
		tenant_id,
		plugin_unique_identifiers,
		source,
		metas,
		func(
			pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
			declaration *plugin_entities.PluginDeclaration,
			meta map[string]any,
		) error {
			runtimeType := plugin_entities.PluginRuntimeType("")

			switch config.Platform {
			case app.PLATFORM_SERVERLESS:
				runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS
			case app.PLATFORM_LOCAL:
				runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
			default:
				return fmt.Errorf("unsupported platform: %s", config.Platform)
			}
			_, _, err := curd.InstallPlugin(tenant_id, pluginUniqueIdentifier, runtimeType, declaration, source, meta)
			return err
		},
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
	baseSSEService(func() (*stream.Stream[plugin_manager.PluginInstallResponse], error) {
		pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		)
		if err != nil {
			return nil, err
		}

		plugin, err := db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
		)
		if err != nil {
			return nil, err
		}

		retStream := stream.NewStream[plugin_manager.PluginInstallResponse](128)

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
	tenant_id string,
	source string,
	meta map[string]any,
	original_plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	new_plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	if original_plugin_unique_identifier == new_plugin_unique_identifier {
		return exception.BadRequestError(errors.New("original and new plugin unique identifier are the same")).ToResponse()
	}

	if original_plugin_unique_identifier.PluginID() != new_plugin_unique_identifier.PluginID() {
		return exception.BadRequestError(errors.New("original and new plugin id are different")).ToResponse()
	}

	// uninstall the original plugin
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_unique_identifier", original_plugin_unique_identifier.String()),
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
		tenant_id,
		[]plugin_entities.PluginUniqueIdentifier{new_plugin_unique_identifier},
		source,
		[]map[string]any{meta},
		func(
			pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
			declaration *plugin_entities.PluginDeclaration,
			meta map[string]any,
		) error {
			originalDeclaration, err := helper.CombinedGetPluginDeclaration(
				original_plugin_unique_identifier,
				plugin_entities.PluginRuntimeType(installation.RuntimeType),
			)
			if err != nil {
				return err
			}

			newDeclaration, err := helper.CombinedGetPluginDeclaration(
				new_plugin_unique_identifier,
				plugin_entities.PluginRuntimeType(installation.RuntimeType),
			)
			if err != nil {
				return err
			}
			// uninstall the original plugin
			upgradeResponse, err := curd.UpgradePlugin(
				tenant_id,
				original_plugin_unique_identifier,
				new_plugin_unique_identifier,
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
			pluginInstallationCacheKey := helper.PluginInstallationCacheKey(original_plugin_unique_identifier.PluginID(), tenant_id)
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
