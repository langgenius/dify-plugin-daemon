package service

import (
	"errors"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"gorm.io/gorm"
)

func InstallPluginFromIdentifiers(
	config *app.Config,
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
) *entities.Response {
	var response struct {
		AllInstalled bool   `json:"all_installed"`
		TaskID       string `json:"task_id"`
	}

	// TODO: create installation task and dispatch to workers
	plugins_wait_for_installation := []plugin_entities.PluginUniqueIdentifier{}

	task := &models.InstallTask{
		Status:           models.InstallTaskStatusRunning,
		TenantID:         tenant_id,
		TotalPlugins:     len(plugin_unique_identifiers),
		CompletedPlugins: 0,
		Plugins:          []models.InstallTaskPluginStatus{},
	}

	for i, plugin_unique_identifier := range plugin_unique_identifiers {
		// check if plugin is already installed
		plugin, err := db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
		)

		task.Plugins = append(task.Plugins, models.InstallTaskPluginStatus{
			PluginUniqueIdentifier: plugin_unique_identifier,
			PluginID:               plugin_unique_identifier.PluginID(),
			Status:                 models.InstallTaskStatusPending,
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
				return entities.NewErrorResponse(-500, err.Error())
			}

			task.CompletedPlugins++
			task.Plugins[i].Status = models.InstallTaskStatusSuccess
			task.Plugins[i].Message = "Installed"
			continue
		}

		if err != db.ErrDatabaseNotFound {
			return entities.NewErrorResponse(-500, err.Error())
		}

		plugins_wait_for_installation = append(plugins_wait_for_installation, plugin_unique_identifier)
	}

	if len(plugins_wait_for_installation) == 0 {
		response.AllInstalled = true
		response.TaskID = ""
		return entities.NewSuccessResponse(response)
	}

	err := db.Create(task)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	response.TaskID = task.ID
	manager := plugin_manager.Manager()

	tasks := []func(){}
	for _, plugin_unique_identifier := range plugins_wait_for_installation {
		tasks = append(tasks, func() {
			updateTaskStatus := func(modifier func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus)) {
				if err := db.WithTransaction(func(tx *gorm.DB) error {
					task, err := db.GetOne[models.InstallTask](
						db.WithTransactionContext(tx),
						db.Equal("id", task.ID),
						db.WLock(), // write lock, multiple tasks can't update the same task
					)
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
						return errors.New("plugin status not found")
					}

					modifier(task_pointer, plugin_status)
					return db.Update(task_pointer, tx)
				}); err != nil {
					log.Error("failed to update install task status %s", err.Error())
				}
			}

			pkg_path, err := manager.GetPackagePath(plugin_unique_identifier)
			if err != nil {
				updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
					task.Status = models.InstallTaskStatusFailed
					plugin.Status = models.InstallTaskStatusFailed
					plugin.Message = err.Error()
				})
				return
			}

			updateTaskStatus(func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
				plugin.Status = models.InstallTaskStatusRunning
				plugin.Message = "Installing"
			})

			var stream *stream.Stream[plugin_manager.PluginInstallResponse]
			if config.Platform == app.PLATFORM_AWS_LAMBDA {
				var zip_decoder *decoder.ZipPluginDecoder
				var pkg_file []byte

				pkg_file, err = os.ReadFile(pkg_path)
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
				stream, err = manager.InstallToAWSFromPkg(tenant_id, zip_decoder, source, meta)
			} else if config.Platform == app.PLATFORM_LOCAL {
				stream, err = manager.InstallToLocal(tenant_id, pkg_path, plugin_unique_identifier, source, meta)
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
	_, err = curd.UninstallPlugin(
		tenant_id,
		plugin_unique_identifier,
		installation.ID,
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("Failed to uninstall plugin: %s", err.Error()))
	}

	return entities.NewSuccessResponse(true)
}
