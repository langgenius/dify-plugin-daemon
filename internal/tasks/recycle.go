package tasks

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/mapping"
)

var (
	// reference to running installation tasks
	installingTasks mapping.Map[string, bool]
)

func startTasks(taskIDs []string) {
	for _, taskID := range taskIDs {
		log.Info("start new install task %s", taskID)
		installingTasks.Store(taskID, true)
	}
}

func endTasks(taskIDs []string) {
	for _, taskID := range taskIDs {
		log.Info("install task %s finished", taskID)
		installingTasks.Delete(taskID)
	}
}

// RecycleTasks is a finalizer to update the status of all running installation tasks to failed
// when the daemon is shutting down
func RecycleTasks() error {
	var errs []error
	installingTasks.Range(func(taskId string, _ bool) bool {
		log.Info("updating task %s status to failed", taskId)
		// update task status to failed
		task, err := db.GetOne[models.InstallTask](
			db.Equal("id", taskId),
			db.InArray("status", []any{
				string(models.InstallTaskStatusRunning),
				string(models.InstallTaskStatusPending)},
			),
		)
		if err != nil {
			errs = append(errs, err)
			return true
		}
		task.Status = models.InstallTaskStatusFailed
		for i := range task.Plugins {
			plugin := &task.Plugins[i]
			plugin.Status = models.InstallTaskStatusFailed
			plugin.Message = "An unexpected daemon shutdown occurred"
		}
		err = db.Update(task)
		if err != nil {
			errs = append(errs, err)
		}
		return true
	})
	return errors.Join(errs...)
}
