package service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

func ListPlugins(tenant_id string, page int, page_size int) *entities.Response {
	type installation struct {
		ID                     string                             `json:"id"`
		Name                   string                             `json:"name"`
		PluginID               string                             `json:"plugin_id"`
		TenantID               string                             `json:"tenant_id"`
		PluginUniqueIdentifier string                             `json:"plugin_unique_identifier"`
		EndpointsActive        int                                `json:"endpoints_active"`
		EndpointsSetups        int                                `json:"endpoints_setups"`
		InstallationID         string                             `json:"installation_id"`
		Declaration            *plugin_entities.PluginDeclaration `json:"declaration"`
		RuntimeType            plugin_entities.PluginRuntimeType  `json:"runtime_type"`
		Version                manifest_entities.Version          `json:"version"`
		CreatedAt              time.Time                          `json:"created_at"`
		UpdatedAt              time.Time                          `json:"updated_at"`
		Source                 string                             `json:"source"`
		Meta                   map[string]any                     `json:"meta"`
	}

	pluginInstallations, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	data := make([]installation, 0, len(pluginInstallations))

	for _, plugin_installation := range pluginInstallations {
		pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
			plugin_installation.PluginUniqueIdentifier,
		)
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
			pluginUniqueIdentifier,
			tenant_id,
			plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
		)
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		data = append(data, installation{
			ID:                     plugin_installation.ID,
			Name:                   pluginDeclaration.Name,
			TenantID:               plugin_installation.TenantID,
			PluginID:               pluginUniqueIdentifier.PluginID(),
			PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
			InstallationID:         plugin_installation.ID,
			Declaration:            pluginDeclaration,
			EndpointsSetups:        plugin_installation.EndpointsSetups,
			EndpointsActive:        plugin_installation.EndpointsActive,
			RuntimeType:            plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
			Version:                pluginDeclaration.Version,
			CreatedAt:              plugin_installation.CreatedAt,
			UpdatedAt:              plugin_installation.UpdatedAt,
			Source:                 plugin_installation.Source,
			Meta:                   plugin_installation.Meta,
		})
	}

	return entities.NewSuccessResponse(data)
}

// Using plugin_ids to fetch plugin installations
func BatchFetchPluginInstallationByIDs(tenant_id string, plugin_ids []string) *entities.Response {
	if len(plugin_ids) == 0 {
		return entities.NewSuccessResponse([]models.PluginInstallation{})
	}

	pluginInstallations, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.InArray("plugin_id", strings.Map(plugin_ids, func(id string) any { return id })),
		db.Page(1, 256), // TODO: pagination
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(pluginInstallations)
}

// check which plugin is missing
func FetchMissingPluginInstallations(tenant_id string, plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier) *entities.Response {
	result := make([]plugin_entities.PluginUniqueIdentifier, 0, len(plugin_unique_identifiers))

	if len(plugin_unique_identifiers) == 0 {
		return entities.NewSuccessResponse(result)
	}

	installed, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.InArray(
			"plugin_unique_identifier",
			strings.Map(
				plugin_unique_identifiers,
				func(id plugin_entities.PluginUniqueIdentifier) any {
					return id.String()
				},
			),
		),
		db.Page(1, 256), // TODO: pagination
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	// check which plugin is missing
	for _, pluginUniqueIdentifier := range plugin_unique_identifiers {
		found := false
		for _, installed_plugin := range installed {
			if installed_plugin.PluginUniqueIdentifier == pluginUniqueIdentifier.String() {
				found = true
				break
			}
		}

		if !found {
			result = append(result, pluginUniqueIdentifier)
		}
	}

	return entities.NewSuccessResponse(result)
}

func ListTools(tenant_id string, page int, page_size int) *entities.Response {
	providers, err := db.GetAll[models.ToolInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(providers)
}

func ListModels(tenant_id string, page int, page_size int) *entities.Response {
	providers, err := db.GetAll[models.AIModelInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(providers)
}

func GetTool(tenant_id string, plugin_id string, provider string) *entities.Response {
	// try get tool
	tool, err := db.GetOne[models.ToolInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	if tool.Provider != provider {
		return entities.NewErrorResponse(-404, "tool not found")
	}

	return entities.NewSuccessResponse(tool)
}
