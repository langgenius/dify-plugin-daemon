package install_service

import (
	"encoding/json"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"gorm.io/gorm"
)

func InstallPlugin(
	tenant_id string,
	user_id string,
	runtime plugin_entities.PluginLifetime,
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
	_, err := curd.UninstallPlugin(tenant_id, plugin_unique_identifier, installation_id)
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
func InstallEndpoint(
	plugin_id plugin_entities.PluginUniqueIdentifier,
	installation_id string,
	tenant_id string,
	user_id string,
	name string,
	settings map[string]any,
) (*models.Endpoint, error) {
	settings_json, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	installation := &models.Endpoint{
		HookID:    strings.RandomString(32),
		PluginID:  plugin_id.PluginID(),
		TenantID:  tenant_id,
		UserID:    user_id,
		Name:      name,
		Enabled:   true,
		ExpiredAt: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
		Settings:  string(settings_json),
	}

	if err := db.WithTransaction(func(tx *gorm.DB) error {
		if err := db.Create(&installation, tx); err != nil {
			return err
		}

		return db.Run(
			db.WithTransactionContext(tx),
			db.Model(models.PluginInstallation{}),
			db.Equal("plugin_id", installation.PluginID),
			db.Equal("tenant_id", installation.TenantID),
			db.Inc(map[string]int{
				"endpoints_setups": 1,
				"endpoints_active": 1,
			}),
		)
	}); err != nil {
		return nil, err
	}

	return installation, nil
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
	return db.WithTransaction(func(tx *gorm.DB) error {
		if err := db.Delete(endpoint, tx); err != nil {
			return err
		}

		// update the plugin installation
		return db.Run(
			db.WithTransactionContext(tx),
			db.Model(models.PluginInstallation{}),
			db.Equal("plugin_id", endpoint.PluginID),
			db.Equal("tenant_id", endpoint.TenantID),
			db.Dec(map[string]int{
				"endpoints_active": 1,
				"endpoints_setups": 1,
			}),
		)
	})
}

func EnabledEndpoint(endpoint *models.Endpoint) error {
	if endpoint.Enabled {
		return nil
	}

	return db.WithTransaction(func(tx *gorm.DB) error {
		endpoint.Enabled = true
		if err := db.Update(endpoint, tx); err != nil {
			return err
		}

		// update the plugin installation
		return db.Run(
			db.WithTransactionContext(tx),
			db.Model(models.PluginInstallation{}),
			db.Equal("plugin_id", endpoint.PluginID),
			db.Equal("tenant_id", endpoint.TenantID),
			db.Inc(map[string]int{
				"endpoints_active": 1,
			}),
		)
	})
}

func DisabledEndpoint(endpoint *models.Endpoint) error {
	if !endpoint.Enabled {
		return nil
	}

	return db.WithTransaction(func(tx *gorm.DB) error {
		endpoint.Enabled = false
		if err := db.Update(endpoint, tx); err != nil {
			return err
		}

		// update the plugin installation
		return db.Run(
			db.WithTransactionContext(tx),
			db.Model(models.PluginInstallation{}),
			db.Equal("plugin_id", endpoint.PluginID),
			db.Equal("tenant_id", endpoint.TenantID),
			db.Dec(map[string]int{
				"endpoints_active": 1,
			}),
		)
	})
}

func UpdateEndpoint(endpoint *models.Endpoint, name string, settings map[string]any) error {
	settings_json, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	endpoint.Name = name
	endpoint.Settings = string(settings_json)

	return db.Update(endpoint)
}
