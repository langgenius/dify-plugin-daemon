package tasks

import (
	"errors"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"gorm.io/gorm"
)

type InstallTaskRegistry struct {
	Order []string
	Tasks map[string]*models.InstallTask
}

func (r *InstallTaskRegistry) IDs() []string {
	ids := make([]string, 0, len(r.Order))
	for _, tenantID := range r.Order {
		if task, ok := r.Tasks[tenantID]; ok {
			ids = append(ids, task.ID)
		}
	}
	return ids
}

func (r *InstallTaskRegistry) PrimaryID() string {
	if len(r.Order) == 0 {
		return ""
	}
	if task, ok := r.Tasks[r.Order[0]]; ok {
		return task.ID
	}
	return ""
}

func truncateMessage(message string) string {
	if len(message) > 1024 {
		message = message[:512] + "..." + message[len(message)-512:]
	}
	return message
}

func SetTaskMessageForOnePlugin(
	taskIDs []string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	message string,
) {
	// avoid message to be too long, only keep the first 512 and last 512 characters
	message = truncateMessage(message)

	for _, taskID := range taskIDs {
		if err := UpdateTaskStatus(taskID, pluginUniqueIdentifier, func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
			plugin.Message = message
		}); err != nil {
			log.Error("failed to update task message", "plugin_unique_identifier", pluginUniqueIdentifier.String(), "error", err)
		}
	}
}

func SetTaskStatusForOnePlugin(
	taskIDs []string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	status models.InstallTaskStatus,
	message string,
) {
	// avoid message to be too long, only keep the first 512 and last 512 characters
	message = truncateMessage(message)

	for _, taskID := range taskIDs {
		if err := UpdateTaskStatus(taskID, pluginUniqueIdentifier, func(task *models.InstallTask, plugin *models.InstallTaskPluginStatus) {
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
			log.Error("failed to update task status", "plugin_unique_identifier", pluginUniqueIdentifier.String(), "error", err)
		}
	}
}

// To update task status more elegant, avoid frequent code like lock and unlock
// this method offers a way to update task status with a modifier function
func UpdateTaskStatus(
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

func DeleteTask(taskId string) error {
	return db.DeleteByCondition(models.InstallTask{
		Model: models.Model{
			ID: taskId,
		},
	})
}

// gracefulShutdownTimeout controls how long RemovePluginIfNeeded waits for
// graceful and forceful shutdown to complete. Package-level variable for test override.
var gracefulShutdownTimeout = 30 * time.Second

func RemovePluginIfNeeded(
	manager plugin_manager.PluginShutdownManager,
	originalPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	response *curd.UpgradePluginResponse,
) error {
	shouldCleanup := response.IsOriginalPluginDeleted

	if shouldCleanup && response.DeletedPlugin != nil && response.DeletedPlugin.InstallType == plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL {
		// uninstall plugin from local install bucket first
		// this must happen before shutdown so that the WatchDog safety net
		// (removeUnusedLocalPlugins) can retry cleanup if shutdown fails
		if err := manager.RemoveLocalPlugin(originalPluginUniqueIdentifier); err != nil {
			return errors.Join(err, errors.New("failed to remove plugin from local install bucket"))
		}

		shutdownGracefully := func() bool {
			ch, err := manager.ShutdownLocalPluginGracefully(originalPluginUniqueIdentifier)
			if err != nil {
				// runtime not found in map, it may have been already cleaned up
				return false
			}
			if ch == nil {
				return true
			}

			// wait for graceful shutdown with a timeout
			select {
			case <-ch:
				return true
			case <-time.After(gracefulShutdownTimeout):
				log.Warn("graceful shutdown timed out, trying forceful shutdown",
					"plugin", originalPluginUniqueIdentifier.String())
				return false
			}
		}

		if shutdownGracefully() {
			return nil
		}

		// graceful shutdown failed or timed out, try forceful as fallback
		forceCh, forceErr := manager.ShutdownLocalPluginForcefully(originalPluginUniqueIdentifier)
		if forceErr != nil {
			return errors.Join(
				forceErr,
				errors.New("failed to shutdown plugin forcefully after graceful shutdown failed"),
			)
		}
		if forceCh != nil {
			select {
			case <-forceCh:
				// forceful shutdown completed
			case <-time.After(gracefulShutdownTimeout):
				return errors.New("forceful shutdown timed out")
			}
		}
	}
	return nil
}
