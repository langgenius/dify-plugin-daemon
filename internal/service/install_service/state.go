package install_service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

func InstallPlugin(
	tenant_id string,
	user_id string,
	runtime plugin_entities.PluginRuntimeInterface,
) (*models.Plugin, *models.PluginInstallation, error) {
	identity, err := runtime.Identity()
	if err != nil {
		return nil, nil, err
	}

	plugin, installation, err := curd.CreatePlugin(tenant_id, user_id, identity, runtime.Type())
	if err != nil {
		return nil, nil, err
	}

	return plugin, installation, nil
}

func UninstallPlugin(tenant_id string, installation_id string, runtime plugin_entities.PluginRuntimeInterface) error {
	identity, err := runtime.Identity()
	if err != nil {
		return err
	}

	// delete the plugin from db
	_, err = curd.DeletePlugin(tenant_id, identity, installation_id)
	if err != nil {
		return err
	}

	return nil
}

// setup a plugin to db,
// returns the endpoint id
func SetupEndpoint(plugin_id string, installation_id string, tenant_id string, user_id string) (string, error) {
	installation := &models.Endpoint{
		HookID:               strings.RandomString(32),
		PluginID:             plugin_id,
		TenantID:             tenant_id,
		UserID:               user_id,
		ExpiredAt:            time.Now().Add(time.Hour * 24 * 365 * 10),
		PluginInstallationId: installation_id,
	}

	if err := db.Create(&installation); err != nil {
		return "", err
	}

	return installation.HookID, nil
}

func GetEndpoint(tenant_id string, plugin_id string, installation_id string) (*models.Endpoint, error) {
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
		db.Equal("plugin_installation_id", installation_id),
	)

	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

// uninstalls a plugin from db
func UninstallEndpoint(endpoint *models.Endpoint) error {
	return db.Delete(endpoint)
}
