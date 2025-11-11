package service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/tasks"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func buildTaskStatuses(
	pluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier,
	declarations []*plugin_entities.PluginDeclaration,
) []models.InstallTaskPluginStatus {
	statuses := make([]models.InstallTaskPluginStatus, len(pluginUniqueIdentifiers))
	for i, identifier := range pluginUniqueIdentifiers {
		statuses[i] = models.InstallTaskPluginStatus{
			PluginUniqueIdentifier: identifier,
			PluginID:               identifier.PluginID(),
			Status:                 models.InstallTaskStatusPending,
			Icon:                   declarations[i].Icon,
			IconDark:               declarations[i].IconDark,
			Labels:                 declarations[i].Label,
			Message:                "",
		}
	}
	return statuses
}

func createInstallTasks(
	tenants []string,
	statuses []models.InstallTaskPluginStatus,
) (*tasks.InstallTaskRegistry, error) {
	registry := &tasks.InstallTaskRegistry{
		Order: append([]string{}, tenants...),
		Tasks: make(map[string]*models.InstallTask, len(tenants)),
	}

	for _, tenantID := range tenants {
		statusCopy := make([]models.InstallTaskPluginStatus, len(statuses))
		copy(statusCopy, statuses)

		task := &models.InstallTask{
			Status:           models.InstallTaskStatusRunning,
			TenantID:         tenantID,
			TotalPlugins:     len(statusCopy),
			CompletedPlugins: 0,
			Plugins:          statusCopy,
		}

		if err := db.Create(task); err != nil {
			return nil, err
		}

		registry.Tasks[tenantID] = task
	}

	return registry, nil
}
