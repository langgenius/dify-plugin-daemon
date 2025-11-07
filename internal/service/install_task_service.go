package service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"gorm.io/gorm"
)

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

// To update task status more elegant, avoid frequent code like lock and unlock
// this method offers a way to update task status with a modifier function
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
