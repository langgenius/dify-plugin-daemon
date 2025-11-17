package install_service

import (
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
)

func InstallPlugin(
	tenant_id string,
	user_id string,
	runtime plugin_entities.PluginLifetime,
	source string,
	meta map[string]any,
) (*models.Plugin, *models.PluginInstallation, error) {
	identity, err := runtime.Identity()
	if err != nil {
		return nil, nil, err
	}

	configuration := runtime.Configuration()
	plugin, installation, err := curd.InstallPlugin(
		tenant_id,
		identity,
		runtime.Type(),
		configuration,
		source,
		meta,
	)

	if err != nil {
		return nil, nil, err
	}

	return plugin, installation, nil
}

func UninstallPlugin(
	tenant_id string,
	installation_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	install_type plugin_entities.PluginRuntimeType,
) error {
	// get declaration
	declaration, err := helper.CombinedGetPluginDeclaration(
		plugin_unique_identifier,
		install_type,
	)
	if err != nil {
		return err
	}
	// delete the plugin from db
	_, err = curd.UninstallPlugin(tenant_id, plugin_unique_identifier, installation_id, declaration)
	if err != nil {
		return err
	}

	// delete endpoints if plugin is not installed through remote
	if install_type != plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
		if err := db.DeleteByCondition(models.Endpoint{
			PluginID: plugin_unique_identifier.PluginID(),
			TenantID: tenant_id,
		}); err != nil {
			return err
		}
	}

	return nil
}
