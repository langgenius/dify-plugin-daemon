package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/verifier"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
)

func UploadPluginFromPkg(
	config *app.Config,
	c *gin.Context,
	tenant_id string,
	dify_pkg_file multipart.File,
	verify_signature bool,
) *entities.Response {
	plugin_file, err := io.ReadAll(dify_pkg_file)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_file)
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	if config.ForceVerifyingSignature || verify_signature {
		err := verifier.VerifyPlugin(decoder)
		if err != nil {
			return entities.NewErrorResponse(-500, errors.Join(err, errors.New(
				"plugin verification has been enabled, and the plugin you want to install has a bad signature",
			)).Error())
		}
	}

	manifest, err := decoder.Manifest()
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	return entities.NewSuccessResponse(manifest.Identity())
}

func InstallPluginFromIdentifiers(
	tenant_id string,
	plugin_unique_identifiers []plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
) *entities.Response {
	// TODO: create installation task and dispatch to workers
	for _, plugin_unique_identifier := range plugin_unique_identifiers {
		if err := InstallPluginFromIdentifier(tenant_id, plugin_unique_identifier, source, meta); err != nil {
			return entities.NewErrorResponse(-500, err.Error())
		}
	}

	return entities.NewSuccessResponse(true)
}

func FetchPluginInstallationTasks(
	tenant_id string,
	page int,
	page_size int,
) *entities.Response {
	return nil
}

func FetchPluginInstallationTask(
	tenant_id string,
	task_id string,
) *entities.Response {
	return nil
}

func FetchPluginManifest(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *entities.Response {
	return nil
}

func InstallPluginFromIdentifier(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
) error {
	// TODO: refactor
	// check if identifier exists
	plugin, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return errors.New("plugin not found")
	}
	if err != nil {
		return err
	}

	if plugin.InstallType == plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
		return errors.New("remote plugin not supported")
	}

	declaration := plugin.Declaration
	// install to this workspace
	if _, _, err := curd.InstallPlugin(tenant_id, plugin_unique_identifier, plugin.InstallType, &declaration, source, meta); err != nil {
		return err
	}

	return nil
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
	plugin_installation_id string,
) *entities.Response {
	// Check if the plugin exists for the tenant
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("id", plugin_installation_id),
	)
	if err == db.ErrDatabaseNotFound {
		return entities.NewErrorResponse(-404, "Plugin installation not found for this tenant")
	}
	if err != nil {
		return entities.NewErrorResponse(-500, err.Error())
	}

	// Uninstall the plugin
	_, err = curd.UninstallPlugin(
		tenant_id,
		plugin_entities.PluginUniqueIdentifier(installation.PluginUniqueIdentifier),
		installation.ID,
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("Failed to uninstall plugin: %s", err.Error()))
	}

	return entities.NewSuccessResponse(true)
}
