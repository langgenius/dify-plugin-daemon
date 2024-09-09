package install_service

import (
	"encoding/json"
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

	configuration := runtime.Configuration()
	plugin, installation, err := curd.CreatePlugin(
		tenant_id,
		user_id,
		identity,
		runtime.Type(),
		configuration,
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
	// delete the plugin from db
	_, err := curd.DeletePlugin(tenant_id, plugin_unique_identifier, installation_id)
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

// setup a plugin to db,
// returns the endpoint id
func InstallEndpoint(
	plugin_id plugin_entities.PluginUniqueIdentifier,
	installation_id string,
	tenant_id string,
	user_id string,
	settings map[string]any,
) (string, error) {
	settings_json, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}

	installation := &models.Endpoint{
		HookID:    strings.RandomString(32),
		PluginID:  plugin_id.PluginID(),
		TenantID:  tenant_id,
		UserID:    user_id,
		Enabled:   true,
		ExpiredAt: time.Now().Add(time.Hour * 24 * 365 * 10),
		Settings:  string(settings_json),
	}

	if err := db.Create(&installation); err != nil {
		return "", err
	}

	return installation.HookID, nil
}

func GetEndpoint(
	tenant_id string, plugin_id string, installation_id string,
) (*models.Endpoint, error) {
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

func EnabledEndpoint(endpoint *models.Endpoint) error {
	endpoint.Enabled = true
	return db.Update(endpoint)
}

func DisabledEndpoint(endpoint *models.Endpoint) error {
	endpoint.Enabled = false
	return db.Update(endpoint)
}
