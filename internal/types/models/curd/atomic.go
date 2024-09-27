package curd

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"gorm.io/gorm"
)

// Create plugin for a tenant, create plugin if it has never been created before
// and install it to the tenant, return the plugin and the installation
// if the plugin has been created before, return the plugin which has been created before
func InstallPlugin(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	install_type plugin_entities.PluginRuntimeType,
	declaration *plugin_entities.PluginDeclaration,
) (
	*models.Plugin, *models.PluginInstallation, error,
) {

	var plugin_to_be_returns *models.Plugin
	var installation_to_be_returns *models.PluginInstallation

	err := db.WithTransaction(func(tx *gorm.DB) error {
		p, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
			db.Equal("plugin_id", plugin_unique_identifier.PluginID()),
			db.Equal("install_type", string(install_type)),
			db.WLock(),
		)

		if err == db.ErrDatabaseNotFound {
			plugin := &models.Plugin{
				PluginID:               plugin_unique_identifier.PluginID(),
				PluginUniqueIdentifier: plugin_unique_identifier.String(),
				InstallType:            install_type,
				Refers:                 1,
				Declaration:            *declaration,
			}

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

		// remove exists installation
		if err := db.DeleteByCondition(
			models.PluginInstallation{
				PluginID:    plugin_to_be_returns.PluginID,
				RuntimeType: string(install_type),
			},
			tx,
		); err != nil {
			return err
		}

		installation := &models.PluginInstallation{
			PluginID:               plugin_to_be_returns.PluginID,
			PluginUniqueIdentifier: plugin_to_be_returns.PluginUniqueIdentifier,
			TenantID:               tenant_id,
			RuntimeType:            string(install_type),
		}

		err = db.Create(installation, tx)
		if err != nil {
			return err
		}

		installation_to_be_returns = installation

		// create tool installation
		if declaration.Tool != nil {
			tool_installation := &models.ToolInstallation{
				PluginID:               plugin_to_be_returns.PluginID,
				PluginUniqueIdentifier: plugin_to_be_returns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
				Provider:               declaration.Tool.Identity.Name,
				Declaration:            *declaration.Tool,
			}

			err := db.Create(tool_installation, tx)
			if err != nil {
				return err
			}
		}

		// create model installation
		if declaration.Model != nil {
			model_installation := &models.AIModelInstallation{
				PluginID:               plugin_to_be_returns.PluginID,
				PluginUniqueIdentifier: plugin_to_be_returns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
				Provider:               declaration.Model.Provider,
				Declaration:            *declaration.Model,
			}

			err := db.Create(model_installation, tx)
			if err != nil {
				return err
			}
		}

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
func UninstallPlugin(tenant_id string, plugin_unique_identifier plugin_entities.PluginUniqueIdentifier, installation_id string) (*DeletePluginResponse, error) {
	var plugin_to_be_returns *models.Plugin
	var installation_to_be_returns *models.PluginInstallation

	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("id", installation_id),
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
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
			db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
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
			db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
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

		// delete tool installation
		declaration := p.Declaration
		if declaration.Tool != nil {
			tool_installation := &models.ToolInstallation{
				PluginID:               plugin_to_be_returns.PluginID,
				PluginUniqueIdentifier: plugin_to_be_returns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
			}

			err := db.DeleteByCondition(&tool_installation, tx)
			if err != nil {
				return err
			}
		}

		// delete model installation
		if declaration.Model != nil {
			model_installation := &models.AIModelInstallation{
				PluginID:               plugin_to_be_returns.PluginID,
				PluginUniqueIdentifier: plugin_to_be_returns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
			}

			err := db.DeleteByCondition(&model_installation, tx)
			if err != nil {
				return err
			}
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
