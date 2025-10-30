package install_service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"gorm.io/gorm"
)

// setup a plugin to db,
func InstallEndpoint(
	plugin_id plugin_entities.PluginUniqueIdentifier,
	installation_id string,
	tenant_id string,
	user_id string,
	name string,
	settings map[string]any,
) (*models.Endpoint, error) {
	installation := &models.Endpoint{
		HookID:    strings.RandomLowercaseString(16),
		PluginID:  plugin_id.PluginID(),
		TenantID:  tenant_id,
		UserID:    user_id,
		Name:      name,
		Enabled:   true,
		ExpiredAt: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
		Settings:  settings,
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
func UninstallEndpoint(endpoint_id string, tenant_id string) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	err := db.WithTransaction(func(tx *gorm.DB) error {
		var err error
		endpoint, err = db.GetOne[models.Endpoint](
			db.WithTransactionContext(tx),
			db.Equal("id", endpoint_id),
			db.Equal("tenant_id", tenant_id),
			db.WLock(),
		)
		if err != nil {
			return err
		}

		if err := db.Delete(&endpoint, tx); err != nil {
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
	return &endpoint, err
}

func EnabledEndpoint(endpointID string, tenantID string) (*models.Endpoint, error) {
	var endpoint *models.Endpoint
	err := db.WithTransaction(func(tx *gorm.DB) error {
		e, err := db.GetOne[models.Endpoint](
			db.WithTransactionContext(tx),
			db.Equal("id", endpointID),
			db.Equal("tenant_id", tenantID),
			db.WLock(),
		)
		if err != nil {
			return err
		}

		endpoint = &e

		if endpoint.Enabled {
			return nil
		}

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
	return endpoint, err
}

func DisabledEndpoint(endpointID string, tenantID string) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	err := db.WithTransaction(func(tx *gorm.DB) error {
		var err error
		endpoint, err = db.GetOne[models.Endpoint](
			db.WithTransactionContext(tx),
			db.Equal("id", endpointID),
			db.Equal("tenant_id", tenantID),
			db.WLock(),
		)
		if err != nil {
			return err
		}
		if !endpoint.Enabled {
			return nil
		}

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
	return &endpoint, err
}

func UpdateEndpoint(endpoint *models.Endpoint, name string, settings map[string]any) error {
	endpoint.Name = name
	endpoint.Settings = settings

	return db.Update(endpoint)
}
