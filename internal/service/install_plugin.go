package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"gorm.io/gorm"
)

type InstallPluginResponse struct {
	AllInstalled bool   `json:"all_installed"`
	TaskID       string `json:"task_id"`
}

type InstallPluginOnDoneHandler func(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	declaration *plugin_entities.PluginDeclaration,
) error

func InstallPluginRuntimeToTenant(
	config *app.Config,
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
	onDone InstallPluginOnDoneHandler, // since installing plugin is a async task, we need to call it asynchronously
) (*InstallPluginResponse, error) {
	response := &InstallPluginResponse{}
	pluginsWaitForInstallation := []plugin_entities.PluginUniqueIdentifier{}

	runtimeType := plugin_entities.PluginRuntimeType("")
	if config.Platform == app.PLATFORM_AWS_LAMBDA {
		runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_AWS
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
			tenant_id,
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
			Labels:                 pluginDeclaration.Label,
			Message:                "",
		})

		if err == nil {
			if err := onDone(pluginUniqueIdentifier, pluginDeclaration); err != nil {
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
	for _, pluginUniqueIdentifier := range pluginsWaitForInstallation {
		// copy the variable to avoid race condition
		pluginUniqueIdentifier := pluginUniqueIdentifier

		declaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			tenant_id,
			runtimeType,
		)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, func() {
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

					// delete the task if all plugins are installed successfully,
					// otherwise update the task status
					if successes == len(taskPointer.Plugins) {
						time.AfterFunc(120*time.Second, func() {
							db.Delete(taskPointer, tx)
						})
					} else {
						return db.Update(taskPointer, tx)
					}

					return nil
				}); err != nil {
					log.Error("failed to update install task status %s", err.Error())
				}
			}

			updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				plugin.Status = models.InstallTaskStatusRunning
				plugin.Message = "Installing"
			})

			var stream *stream.Stream[plugin_manager.PluginInstallResponse]
			if config.Platform == app.PLATFORM_AWS_LAMBDA {
				var zipDecoder *decoder.ZipPluginDecoder
				var pkgFile []byte

				pkgFile, err = manager.GetPackage(pluginUniqueIdentifier)
				if err != nil {
					updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
						task.Status = models.InstallTaskStatusFailed
						plugin.Status = models.InstallTaskStatusFailed
						plugin.Message = "Failed to read plugin package"
					})
					return
				}

				zipDecoder, err = decoder.NewZipPluginDecoder(pkgFile)
				if err != nil {
					updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
						task.Status = models.InstallTaskStatusFailed
						plugin.Status = models.InstallTaskStatusFailed
						plugin.Message = err.Error()
					})
					return
				}
				stream, err = manager.InstallToAWSFromPkg(zipDecoder, source, meta)
			} else if config.Platform == app.PLATFORM_LOCAL {
				stream, err = manager.InstallToLocal(pluginUniqueIdentifier, source, meta)
			} else {
				updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					task.Status = models.InstallTaskStatusFailed
					plugin.Status = models.InstallTaskStatusFailed
					plugin.Message = "Unsupported platform"
				})
				return
			}

			if err != nil {
				updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					task.Status = models.InstallTaskStatusFailed
					plugin.Status = models.InstallTaskStatusFailed
					plugin.Message = err.Error()
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

				if message.Event == plugin_manager.PluginInstallEventError {
					updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
						task.Status = models.InstallTaskStatusFailed
						plugin.Status = models.InstallTaskStatusFailed
						plugin.Message = message.Data
					})
					return
				}

				if message.Event == plugin_manager.PluginInstallEventDone {
					if err := onDone(pluginUniqueIdentifier, declaration); err != nil {
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
		})
	}

	// submit async tasks
	routine.WithMaxRoutine(3, tasks)

	return response, nil
}

func InstallPluginFromIdentifiers(
	config *app.Config,
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
) *entities.Response {
	response, err := InstallPluginRuntimeToTenant(
		config,
		tenant_id,
		plugin_unique_identifiers,
		source,
		meta,
		func(
			pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
			declaration *plugin_entities.PluginDeclaration,
		) error {
			runtimeType := plugin_entities.PluginRuntimeType("")

			switch config.Platform {
			case app.PLATFORM_AWS_LAMBDA:
				runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_AWS
			case app.PLATFORM_LOCAL:
				runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
			default:
				return fmt.Errorf("unsupported platform: %s", config.Platform)
			}

			_, _, err := curd.InstallPlugin(
				tenant_id,
				pluginUniqueIdentifier,
				runtimeType,
				declaration,
				source,
				meta,
			)
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
		meta,
		func(
			pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
			declaration *plugin_entities.PluginDeclaration,
		) error {
			// uninstall the original plugin
			upgradeResponse, err := curd.UpgradePlugin(
				tenant_id,
				original_plugin_unique_identifier,
				new_plugin_unique_identifier,
				declaration,
				plugin_entities.PluginRuntimeType(installation.RuntimeType),
				source,
				meta,
			)

			if err != nil {
				return err
			}

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
		return exception.PluginUniqueIdentifierError(err).ToResponse()
	}

	// Uninstall the plugin
	deleteResponse, err := curd.UninstallPlugin(
		tenant_id,
		pluginUniqueIdentifier,
		installation.ID,
	)
	if err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to uninstall plugin: %s", err.Error())).ToResponse()
	}

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
