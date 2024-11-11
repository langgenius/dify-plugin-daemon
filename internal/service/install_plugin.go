package service

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
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
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	declaration *plugin_entities.PluginDeclaration,
) error

func InstallPluginRuntimeToTenant(
	config *app.Config,
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
	on_done InstallPluginOnDoneHandler, // since installing plugin is a async task, we need to call it asynchronously
) (*InstallPluginResponse, error) {
	response := &InstallPluginResponse{}

	plugins_wait_for_installation := []plugin_entities.PluginUniqueIdentifier{}

	task := &models.InstallTask{
		Status:           models.InstallTaskStatusRunning,
		TenantID:         tenant_id,
		TotalPlugins:     len(plugin_unique_identifiers),
		CompletedPlugins: 0,
		Plugins:          []models.InstallTaskPluginStatus{},
	}

	for i, plugin_unique_identifier := range plugin_unique_identifiers {
		// fetch plugin declaration first, before installing, we need to ensure pkg is uploaded
		plugin_declaration, err := helper.CombinedGetPluginDeclaration(plugin_unique_identifier)
		if err != nil {
			return nil, err
		}

		// check if plugin is already installed
		plugin, err := db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
		)

		task.Plugins = append(task.Plugins, models.InstallTaskPluginStatus{
			PluginUniqueIdentifier: plugin_unique_identifier,
			PluginID:               plugin_unique_identifier.PluginID(),
			Status:                 models.InstallTaskStatusPending,
			Icon:                   plugin_declaration.Icon,
			Labels:                 plugin_declaration.Label,
			Message:                "",
		})

		if err == nil {
			// already installed by other tenant
			declaration := plugin.Declaration
			if _, _, err := curd.InstallPlugin(
				tenant_id,
				plugin_unique_identifier,
				plugin.InstallType,
				&declaration,
				source,
				meta,
			); err != nil {
				return nil, err
			}

			task.CompletedPlugins++
			task.Plugins[i].Status = models.InstallTaskStatusSuccess
			task.Plugins[i].Message = "Installed"
			continue
		}

		if err != db.ErrDatabaseNotFound {
			return nil, err
		}

		plugins_wait_for_installation = append(plugins_wait_for_installation, plugin_unique_identifier)
	}

	if len(plugins_wait_for_installation) == 0 {
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
	for _, plugin_unique_identifier := range plugins_wait_for_installation {
		// copy the variable to avoid race condition
		plugin_unique_identifier := plugin_unique_identifier

		declaration, err := manager.GetDeclaration(plugin_unique_identifier)
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

					task_pointer := &task
					var plugin_status *models.InstallTaskPluginStatus
					for i := range task.Plugins {
						if task.Plugins[i].PluginUniqueIdentifier == plugin_unique_identifier {
							plugin_status = &task.Plugins[i]
							break
						}
					}

					if plugin_status == nil {
						return nil
					}

					modifier(task_pointer, plugin_status)
					return db.Update(task_pointer, tx)
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
				var zip_decoder *decoder.ZipPluginDecoder
				var pkg_file []byte

				pkg_file, err = manager.GetPackage(plugin_unique_identifier)
				if err != nil {
					updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
						task.Status = models.InstallTaskStatusFailed
						plugin.Status = models.InstallTaskStatusFailed
						plugin.Message = "Failed to read plugin package"
					})
					return
				}

				zip_decoder, err = decoder.NewZipPluginDecoder(pkg_file)
				if err != nil {
					updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
						task.Status = models.InstallTaskStatusFailed
						plugin.Status = models.InstallTaskStatusFailed
						plugin.Message = err.Error()
					})
					return
				}
				stream, err = manager.InstallToAWSFromPkg(zip_decoder, source, meta)
			} else if config.Platform == app.PLATFORM_LOCAL {
				stream, err = manager.InstallToLocal(plugin_unique_identifier, source, meta)
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
					if err := on_done(plugin_unique_identifier, declaration); err != nil {
						updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
							task.Status = models.InstallTaskStatusFailed
							plugin.Status = models.InstallTaskStatusFailed
							plugin.Message = "Failed to create plugin"
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
	response, err := InstallPluginRuntimeToTenant(config, tenant_id, plugin_unique_identifiers, source, meta, func(
		plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
		declaration *plugin_entities.PluginDeclaration,
	) error {
		runtime_type := plugin_entities.PluginRuntimeType("")

		switch config.Platform {
		case app.PLATFORM_AWS_LAMBDA:
			runtime_type = plugin_entities.PLUGIN_RUNTIME_TYPE_AWS
		case app.PLATFORM_LOCAL:
			runtime_type = plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
		default:
			return fmt.Errorf("unsupported platform: %s", config.Platform)
		}

		_, _, err := curd.InstallPlugin(
			tenant_id,
			plugin_unique_identifier,
			runtime_type,
			declaration,
			source,
			meta,
		)
		return err
	})
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
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
		return entities.NewErrorResponse(-400, "original and new plugin unique identifier are the same")
	}

	if original_plugin_unique_identifier.PluginID() != new_plugin_unique_identifier.PluginID() {
		return entities.NewErrorResponse(-400, "original and new plugin id are different")
	}

	// uninstall the original plugin
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_unique_identifier", original_plugin_unique_identifier.String()),
		db.Equal("source", source),
	)

	if err == db.ErrDatabaseNotFound {
		return entities.NewErrorResponse(-404, "Plugin installation not found for this tenant")
	}

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	// install the new plugin runtime
	response, err := InstallPluginRuntimeToTenant(
		config,
		tenant_id,
		[]plugin_entities.PluginUniqueIdentifier{new_plugin_unique_identifier},
		source,
		meta,
		func(
			plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
			declaration *plugin_entities.PluginDeclaration,
		) error {
			// uninstall the original plugin
			err = curd.UpgradePlugin(
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

			return nil
		},
	)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
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
		return entities.NewErrorResponse(-500, err.Error())
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
		return entities.NewErrorResponse(-500, err.Error())
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
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(true)
}

func DeletePluginInstallationItemFromTask(
	tenant_id string,
	task_id string,
	identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	item, err := db.GetOne[models.InstallTask](
		db.Equal("task_id", task_id),
		db.Equal("tenant_id", tenant_id),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	plugins := []models.InstallTaskPluginStatus{}
	for _, plugin := range item.Plugins {
		if plugin.PluginUniqueIdentifier != identifier {
			plugins = append(plugins, plugin)
		}
	}

	if len(plugins) == 0 {
		err = db.Delete(&item)
	} else {
		item.Plugins = plugins
		err = db.Update(&item)
	}

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(true)
}

func FetchPluginFromIdentifier(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	_, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewSuccessResponse(false)
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
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
		return entities.NewErrorResponse(-404, "Plugin installation not found for this tenant")
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	plugin_unique_identifier, err := plugin_entities.NewPluginUniqueIdentifier(installation.PluginUniqueIdentifier)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to parse plugin unique identifier: %v", err))
	}

	// Uninstall the plugin
	delete_response, err := curd.UninstallPlugin(
		tenant_id,
		plugin_unique_identifier,
		installation.ID,
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("Failed to uninstall plugin: %s", err.Error()))
	}

	if delete_response.IsPluginDeleted {
		// delete the plugin if no installation left
		manager := plugin_manager.Manager()
		if delete_response.Installation.RuntimeType == string(
			plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		) {
			err = manager.UninstallFromLocal(plugin_unique_identifier)
			if err != nil {
				return entities.NewErrorResponse(-500, fmt.Sprintf("Failed to uninstall plugin: %s", err.Error()))
			}
		}
	}

	return entities.NewSuccessResponse(true)
}
