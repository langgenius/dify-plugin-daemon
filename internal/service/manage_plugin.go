package service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
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
		plugin, err := cache.Get[models.Plugin](plugin_installation.PluginUniqueIdentifier)
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		declaration, err := plugin.GetDeclaration()
		if err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}

		data = append(data, installation{
			ID:             plugin_installation.ID,
			Name:           declaration.Name,
			PluginID:       plugin.ID,
			InstallationID: plugin_installation.ID,
			Description:    declaration,
			RuntimeType:    plugin_entities.PluginRuntimeType(plugin_installation.RuntimeType),
			Version:        declaration.Version,
			CreatedAt:      plugin_installation.CreatedAt,
			UpdatedAt:      plugin_installation.UpdatedAt,
		})

	}

	return entities.NewSuccessResponse(data)
}

func ListTools(tenant_id string, page int, page_size int) *entities.Response {
	return nil
}

func ListModels(tenant_id string, page int, page_size int) *entities.Response {
	return nil
}
