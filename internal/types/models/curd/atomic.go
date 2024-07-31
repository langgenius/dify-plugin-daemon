package curd

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"gorm.io/gorm"
)

// Create plugin for a tenant, create plugin if it has never been created before
// and install it to the tenant, return the plugin and the installation
// if the plugin has been created before, return the plugin which has been created before
func CreatePlugin(tenant_id string, user_id string, plugin *models.Plugin) (*models.Plugin, *models.PluginInstallation, error) {
	var plugin_to_be_returns *models.Plugin
	var installation_to_be_returns *models.PluginInstallation

	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", plugin_to_be_returns.PluginID),
		db.Equal("tenant_id", tenant_id),
	)

	if err != nil && err != db.ErrDatabaseNotFound {
		return nil, nil, err
	} else if err != nil {
		return nil, nil, errors.New("plugin has been installed already")
	}

	err = db.WithTransaction(func(tx *gorm.DB) error {
		p, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_id", plugin.PluginID),
			db.WLock(),
		)

		if err == db.ErrDatabaseNotFound {
			plugin.Refers = 1
			err := db.Create(plugin, tx)
			if err != nil {
				return err
			}

			plugin_to_be_returns = plugin
		} else if err != nil {
			return err
		} else {
			p.Refers++
			err := db.Update(&p, tx)
			if err != nil {
				return err
			}
			plugin_to_be_returns = &p
		}

		installation := &models.PluginInstallation{
			PluginID: plugin_to_be_returns.PluginID,
			TenantID: tenant_id,
			UserID:   user_id,
		}

		err = db.Create(installation, tx)
		if err != nil {
			return err
		}

		installation_to_be_returns = installation

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return plugin_to_be_returns, installation_to_be_returns, nil
}

type DeletePluginResponse struct {
	Plugin          *models.Plugin
	Installation    *models.PluginInstallation
	IsPluginDeleted bool
}

// Delete plugin for a tenant, delete the plugin if it has never been created before
// and uninstall it from the tenant, return the plugin and the installation
// if the plugin has been created before, return the plugin which has been created before
func DeletePlugin(tenant_id string, plugin_id string) (*DeletePluginResponse, error) {
	var plugin_to_be_returns *models.Plugin
	var installation_to_be_returns *models.PluginInstallation

	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", plugin_to_be_returns.PluginID),
		db.Equal("tenant_id", tenant_id),
	)

	if err != nil {
		if err == db.ErrDatabaseNotFound {
			return nil, errors.New("plugin has not been installed")
		} else {
			return nil, err
		}
	}

	err = db.WithTransaction(func(tx *gorm.DB) error {
		p, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_id", plugin_to_be_returns.PluginID),
			db.WLock(),
		)

		if err == db.ErrDatabaseNotFound {
			return errors.New("plugin has not been installed")
		} else if err != nil {
			return err
		} else {
			p.Refers--
			err := db.Update(&p, tx)
			if err != nil {
				return err
			}
			plugin_to_be_returns = &p
		}

		installation, err := db.GetOne[models.PluginInstallation](
			db.WithTransactionContext(tx),
			db.Equal("plugin_id", plugin_id),
			db.Equal("tenant_id", tenant_id),
		)

		if err == db.ErrDatabaseNotFound {
			return errors.New("plugin has not been installed")
		} else if err != nil {
			return err
		} else {
			err := db.Delete(&installation, tx)
			if err != nil {
				return err
			}
			installation_to_be_returns = &installation
		}

		if plugin_to_be_returns.Refers == 0 {
			err := db.Delete(&plugin_to_be_returns, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &DeletePluginResponse{
		Plugin:          plugin_to_be_returns,
		Installation:    installation_to_be_returns,
		IsPluginDeleted: plugin_to_be_returns.Refers == 0,
	}, nil
}
