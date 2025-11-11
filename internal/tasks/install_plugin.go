package tasks

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type PluginInstallJob struct {
	Identifier          plugin_entities.PluginUniqueIdentifier
	Declaration         *plugin_entities.PluginDeclaration
	Meta                map[string]any
	NeedsRuntimeInstall bool
}

func ProcessInstallJob(
	manager *plugin_manager.PluginManager,
	tenants []string,
	runtimeType plugin_entities.PluginRuntimeType,
	source string,
	taskIDs []string,
	job PluginInstallJob,
) {
	// if the plugin does not need runtime install, just save the installation to the database
	if !job.NeedsRuntimeInstall {
		if err := SaveInstallationForTenantsToDB(tenants, job, runtimeType, source); err != nil {
			SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, err.Error())
			return
		}
		SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusSuccess, "Installed")
		return
	}

	// start installation process
	installationStream, err := manager.Install(job.Identifier)
	if err != nil {
		SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, fmt.Sprintf("failed to start installation: %v", err))
		return
	}

	// wait for the job to be done
	err = installationStream.Async(func(resp installation_entities.PluginInstallResponse) {
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
			SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusSuccess, "Installed")
		}
	})
	if err != nil {
		SetTaskStatusForOnePlugin(taskIDs, job.Identifier, models.InstallTaskStatusFailed, err.Error())
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
