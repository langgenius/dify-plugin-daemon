package service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
)

func ListPlugins(tenant_id string, page int, page_size int) *entities.Response {
	type installation struct {
		ID             string                             `json:"id"`
		Name           string                             `json:"name"`
		PluginID       string                             `json:"plugin_id"`
		InstallationID string                             `json:"installation_id"`
		Description    *plugin_entities.PluginDeclaration `json:"description"`
		RuntimeType    plugin_entities.PluginRuntimeType  `json:"runtime_type"`
		Version        string                             `json:"version"`
		CreatedAt      time.Time                          `json:"created_at"`
		UpdatedAt      time.Time                          `json:"updated_at"`
	}

	plugin_installations, err := db.GetAll[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Page(page, page_size),
	)

	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	data := make([]installation, 0, len(plugin_installations))

	for _, plugin_installation := range plugin_installations {
		plugin_unique_identifier, err := plugin_entities.NewPluginUniqueIdentifier(
			plugin_installation.PluginUniqueIdentifier,
		)
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		plugin_declaration, err := helper.CombinedGetPluginDeclaration(plugin_unique_identifier)
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		data = append(data, installation{
			ID:             plugin_installation.ID,
			Name:           plugin_declaration.Name,
			PluginID:       plugin_unique_identifier.PluginID(),
			InstallationID: plugin_installation.ID,
			Description:    plugin_declaration,
			RuntimeType:    plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
			Version:        plugin_declaration.Version,
			CreatedAt:      plugin_installation.CreatedAt,
			UpdatedAt:      plugin_installation.UpdatedAt,
		})
	}

	return entities.NewSuccessResponse(data)
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
