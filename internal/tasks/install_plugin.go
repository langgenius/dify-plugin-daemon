package tasks

import (
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type PluginInstallJob struct {
	Identifier          plugin_entities.PluginUniqueIdentifier
	Declaration         *plugin_entities.PluginDeclaration
	Meta                map[string]any
	NeedsRuntimeInstall bool
}

type PluginUpgradeJob struct {
	NewIdentifier       plugin_entities.PluginUniqueIdentifier
	NewDeclaration      *plugin_entities.PluginDeclaration
	OriginalIdentifier  plugin_entities.PluginUniqueIdentifier
	OriginalDeclaration *plugin_entities.PluginDeclaration
	Meta                map[string]any
}

func ProcessInstallJob(
	manager *plugin_manager.PluginManager,
	tenants []string,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
	taskIDs []string,
	job PluginInstallJob,
) {
	startTasks(taskIDs)
	defer endTasks(taskIDs)

	// if the plugin does not need runtime install, just save the installation to the database
	if !job.NeedsRuntimeInstall {
		if err := SaveInstallationForTenantsToDB(tenants, job, runtimeType, source); err != nil {
			SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, err.Error())
			return
		}
		SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusSuccess, "installed")
		return
	}

	// set status to running
	SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusRunning, "starting")

	// start installation process
	installationStream, err := manager.Install(job.Identifier)
	if err != nil {
		SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, fmt.Sprintf("failed to start installation: %v", err))
		return
	}

	// wait for the job to be done
	err = installationStream.Process(func(resp installation_entities.PluginInstallResponse) {
		switch resp.Event {
		case installation_entities.PluginInstallEventInfo:
			SetTaskMessageForOnePlugin(taskIDs, job.Identifier, resp.Data)
		case installation_entities.PluginInstallEventError:
			SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, resp.Data)
		case installation_entities.PluginInstallEventDone:
			if err := SaveInstallationForTenantsToDB(tenants, job, runtimeType, source); err != nil {
				SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, err.Error())
				return
			}
			SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusSuccess, "installed")
			// delete the task in 60 seconds
			time.AfterFunc(time.Second*60, func() {
				for _, taskID := range taskIDs {
					if err := DeleteTask(taskID); err != nil {
						log.Error("failed to delete task %s: %v", taskID, err)
					}
				}
			})
		}
	})
	if err != nil {
		SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, err.Error())
	}
}

func ProcessUpgradeJob(
	manager *plugin_manager.PluginManager,
	tenants []string,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
	taskIDs []string,
	job PluginUpgradeJob,
) {
	startTasks(taskIDs)
	defer endTasks(taskIDs)

	// set status to running
	SetTaskStatusForOnePlugin(taskIDs, job.NewIdentifier, models.InstallTaskStatusRunning, "starting")

	// start installation process
	installationStream, err := manager.Install(job.NewIdentifier)
	if err != nil {
		SetTaskStatusForOnePlugin(taskIDs, job.NewIdentifier, models.InstallTaskStatusFailed, fmt.Sprintf("failed to start installation: %v", err))
		return
	}

	err = installationStream.Process(func(resp installation_entities.PluginInstallResponse) {
		switch resp.Event {
		case installation_entities.PluginInstallEventInfo:
			SetTaskMessageForOnePlugin(taskIDs, job.NewIdentifier, resp.Data)
		case installation_entities.PluginInstallEventError:
			SetTaskStatusForOnePlugin(taskIDs, job.NewIdentifier, models.InstallTaskStatusFailed, resp.Data)
		case installation_entities.PluginInstallEventDone:
			for _, tenantID := range tenants {
				response, err := curd.UpgradePlugin(
					tenantID,
					job.OriginalIdentifier,
					job.NewIdentifier,
					job.OriginalDeclaration,
					job.NewDeclaration,
					runtimeType,
					source,
					job.Meta,
				)
				if err != nil {
					SetTaskStatusForOnePlugin(taskIDs, job.NewIdentifier, models.InstallTaskStatusFailed, err.Error())
					return
				}

				if err := RemovePluginIfNeeded(manager, job.OriginalIdentifier, response); err != nil {
					log.Error("failed to remove uninstalled plugin: %v", err)
				}
			}

			SetTaskStatusForOnePlugin(taskIDs, job.NewIdentifier, models.InstallTaskStatusSuccess, "upgraded")
		}
	})
	if err != nil {
		SetTaskStatusForOnePlugin(taskIDs, job.NewIdentifier, models.InstallTaskStatusFailed, err.Error())
	}

}

func SaveInstallationForTenantsToDB(
	tenants []string,
	job PluginInstallJob,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
) error {
	for _, tenantID := range tenants {
		if err := SaveInstallationForTenantToDB(tenantID, job, runtimeType, source); err != nil {
			return err
		}
	}
	return nil
}

func SaveInstallationForTenantToDB(
	tenantID string,
	job PluginInstallJob,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
) error {
	_, _, err := curd.InstallPlugin(
		tenantID,
		job.Identifier,
		runtimeType,
		job.Declaration,
		source,
		job.Meta,
	)
	if err != nil && err != curd.ErrPluginAlreadyInstalled {
		return err
	}
	return nil
}
