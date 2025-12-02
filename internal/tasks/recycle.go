package tasks

import (
	"errors"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
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

func markTasksAsTimeout(tasks []*models.InstallTask) {
	if len(tasks) == 0 {
		return
	}
	for _, task := range tasks {
		task.Status = models.InstallTaskStatusFailed
		for i := range task.Plugins {
			plugin := &task.Plugins[i]
			plugin.Status = models.InstallTaskStatusFailed
			plugin.Message = "Task timed out but not properly terminated"
		}
	}
	db.Update(tasks)
}

// Just in case some tasks may stuck for some reason that we don't know.
func MonitorTimeoutTasks(cluster *cluster.Cluster, config *app.Config) {
	go func() {

		var timeout time.Duration
		if config.Platform == app.PLATFORM_SERVERLESS {
			timeout = time.Duration(config.DifyPluginServerlessConnectorLaunchTimeout) * time.Second
		} else {
			timeout = time.Duration(config.PythonEnvInitTimeout) * time.Second
		}
		// add some tolerance to timeout to avoid race condition
		timeout = timeout + time.Minute

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if !cluster.IsMaster() {
				continue
			}
			tasksToProcess := []*models.InstallTask{}
			// get all tasks that are pending or running
			tasks, err := db.GetAll[models.InstallTask](
				db.InArray("status", []any{
					string(models.InstallTaskStatusPending),
					string(models.InstallTaskStatusRunning),
				}),
			)
			if err != nil {
				log.Error("failed to get all tasks: %v", err)
				continue
			}
			for i := range tasks {
				task := &tasks[i]
				if time.Since(task.CreatedAt) > timeout {
					log.Info("task %s timed out", task.ID)
					tasksToProcess = append(tasksToProcess, task)
				}
			}
			markTasksAsTimeout(tasksToProcess)
		}
	}()

}
