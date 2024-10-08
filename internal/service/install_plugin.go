package service

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InstallPluginFromPkg(c *gin.Context, tenant_id string, dify_pkg_file multipart.File) {
	manager := plugin_manager.Manager()

	plugin_file, err := io.ReadAll(dify_pkg_file)
	if err != nil {
		c.JSON(200, entities.NewErrorResponse(-500, err.Error()))
		return
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_file)
	if err != nil {
		c.JSON(200, entities.NewErrorResponse(-500, err.Error()))
		return
	}

	baseSSEService(
		func() (*stream.Stream[plugin_manager.PluginInstallResponse], error) {
			return manager.Install(tenant_id, decoder)
		},
		c,
		3600,
	)
}

func InstallPluginFromIdentifier(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	// check if identifier exists
	plugin, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewErrorResponse(-404, "Plugin not found")
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	if plugin.InstallType == plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
		return entities.NewErrorResponse(-500, "remote plugin not supported")
	}

	declaration := plugin.Declaration
	// install to this workspace
	if _, _, err := curd.InstallPlugin(tenant_id, plugin_unique_identifier, plugin.InstallType, &declaration); err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(true)
}

func FetchPluginFromIdentifier(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	_, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewSuccessResponse(false)
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(true)
}

func UninstallPlugin(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	// Check if the plugin exists for the tenant
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewErrorResponse(-404, "Plugin not found for this tenant")
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	// Uninstall the plugin
	_, err = curd.UninstallPlugin(tenant_id, plugin_unique_identifier, installation.ID)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("Failed to uninstall plugin: %s", err.Error()))
	}

	return entities.NewSuccessResponse(true)
}
