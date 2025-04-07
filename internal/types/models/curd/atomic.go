package curd

import (
	"errors"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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
	source string,
	meta map[string]any,
) (
	*models.Plugin, *models.PluginInstallation, error,
) {

	var pluginToBeReturns *models.Plugin
	var installationToBeReturns *models.PluginInstallation

	// check if already installed
	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", plugin_unique_identifier.PluginID()),
		db.Equal("tenant_id", tenant_id),
	)

	if err == nil {
		return nil, nil, ErrPluginAlreadyInstalled
	}

	err = db.WithTransaction(func(tx *gorm.DB) error {
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
			}

			if install_type == plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
				plugin.RemoteDeclaration = *declaration
			}

			err := db.Create(plugin, tx)
			if err != nil {
				return err
			}

			pluginToBeReturns = plugin
		} else if err != nil {
			return err
		} else {
			p.Refers++
			err := db.Update(&p, tx)
			if err != nil {
				return err
			}
			pluginToBeReturns = &p
		}

		// remove exists installation
		if err := db.DeleteByCondition(
			models.PluginInstallation{
				PluginID:    pluginToBeReturns.PluginID,
				RuntimeType: string(install_type),
				TenantID:    tenant_id,
			},
			tx,
		); err != nil {
			return err
		}

		installation := &models.PluginInstallation{
			PluginID:               pluginToBeReturns.PluginID,
			PluginUniqueIdentifier: pluginToBeReturns.PluginUniqueIdentifier,
			TenantID:               tenant_id,
			RuntimeType:            string(install_type),
			Source:                 source,
			Meta:                   meta,
		}

		err = db.Create(installation, tx)
		if err != nil {
			return err
		}

		installationToBeReturns = installation

		// create tool installation
		if declaration.Tool != nil {
			toolInstallation := &models.ToolInstallation{
				PluginID:               pluginToBeReturns.PluginID,
				PluginUniqueIdentifier: pluginToBeReturns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
				Provider:               declaration.Tool.Identity.Name,
			}

			err := db.Create(toolInstallation, tx)
			if err != nil {
				return err
			}
		}

		// create agent installation
		if declaration.AgentStrategy != nil {
			agentStrategyInstallation := &models.AgentStrategyInstallation{
				PluginID:               pluginToBeReturns.PluginID,
				PluginUniqueIdentifier: pluginToBeReturns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
				Provider:               declaration.AgentStrategy.Identity.Name,
			}

			err := db.Create(agentStrategyInstallation, tx)
			if err != nil {
				return err
			}
		}

		// create model installation
		if declaration.Model != nil {
			modelInstallation := &models.AIModelInstallation{
				PluginID:               pluginToBeReturns.PluginID,
				PluginUniqueIdentifier: pluginToBeReturns.PluginUniqueIdentifier,
				TenantID:               tenant_id,
				Provider:               declaration.Model.Provider,
			}

			err := db.Create(modelInstallation, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return pluginToBeReturns, installationToBeReturns, nil
}

type DeletePluginResponse struct {
	Plugin       *models.Plugin
	Installation *models.PluginInstallation

	// whether the refers of the plugin has been decreased to 0
	// which means the whole plugin has been uninstalled, not just the installation
	IsPluginDeleted bool
}

// Delete plugin for a tenant, delete the plugin if it has never been created before
// and uninstall it from the tenant, return the plugin and the installation
// if the plugin has been created before, return the plugin which has been created before
func UninstallPlugin(
	tenant_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	installation_id string,
	declaration *plugin_entities.PluginDeclaration,
) (*DeletePluginResponse, error) {
	var pluginToBeReturns *models.Plugin
	var installationToBeReturns *models.PluginInstallation

	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("id", installation_id),
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
		db.Equal("tenant_id", tenant_id),
	)

	pluginInstallationCacheKey := strings.Join(
		[]string{
			"plugin_id",
			plugin_unique_identifier.PluginID(),
			"tenant_id",
			tenant_id,
		},
		":",
	)

	_ = cache.AutoDelete[models.PluginInstallation](pluginInstallationCacheKey)

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
			pluginToBeReturns = &p
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
			installationToBeReturns = &installation
		}

		// delete tool installation
		if declaration.Tool != nil {
			toolInstallation := &models.ToolInstallation{
				PluginID: pluginToBeReturns.PluginID,
				TenantID: tenant_id,
			}

			err := db.DeleteByCondition(&toolInstallation, tx)
			if err != nil {
				return err
			}
		}

		// delete agent installation
		if declaration.AgentStrategy != nil {
			agentStrategyInstallation := &models.AgentStrategyInstallation{
				PluginID: pluginToBeReturns.PluginID,
				TenantID: tenant_id,
			}

			err := db.DeleteByCondition(&agentStrategyInstallation, tx)
			if err != nil {
				return err
			}
		}

		// delete model installation
		if declaration.Model != nil {
			modelInstallation := &models.AIModelInstallation{
				PluginID: pluginToBeReturns.PluginID,
				TenantID: tenant_id,
			}

			err := db.DeleteByCondition(&modelInstallation, tx)
			if err != nil {
				return err
			}
		}

		if pluginToBeReturns.Refers == 0 {
			err := db.Delete(&pluginToBeReturns, tx)
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
		Plugin:          pluginToBeReturns,
		Installation:    installationToBeReturns,
		IsPluginDeleted: pluginToBeReturns.Refers == 0,
	}, nil
}

type UpgradePluginResponse struct {
	// whether the original plugin has been deleted
	IsOriginalPluginDeleted bool

	// the deleted plugin
	DeletedPlugin *models.Plugin
}

// Upgrade plugin for a tenant, upgrade the plugin if it has been created before
// and uninstall the original plugin and install the new plugin, but keep the original installation information
// like endpoint_setups, etc.
func UpgradePlugin(
	tenant_id string,
	original_plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	new_plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	originalDeclaration *plugin_entities.PluginDeclaration,
	newDeclaration *plugin_entities.PluginDeclaration,
	install_type plugin_entities.PluginRuntimeType,
	source string,
	meta map[string]any,
) (*UpgradePluginResponse, error) {
	var response UpgradePluginResponse

	err := db.WithTransaction(func(tx *gorm.DB) error {
		installation, err := db.GetOne[models.PluginInstallation](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", original_plugin_unique_identifier.String()),
			db.Equal("tenant_id", tenant_id),
			db.WLock(),
		)

		if err == db.ErrDatabaseNotFound {
			return errors.New("plugin has not been installed")
		} else if err != nil {
			return err
		}

		// check if the new plugin has existed
		plugin, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", new_plugin_unique_identifier.String()),
		)

		if err == db.ErrDatabaseNotFound {
			// create new plugin
			plugin = models.Plugin{
				PluginID:               new_plugin_unique_identifier.PluginID(),
				PluginUniqueIdentifier: new_plugin_unique_identifier.String(),
				InstallType:            install_type,
				Refers:                 0,
				ManifestType:           manifest_entities.PluginType,
			}

			err := db.Create(&plugin, tx)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// update exists installation
		installation.PluginUniqueIdentifier = new_plugin_unique_identifier.String()
		installation.Meta = meta
		err = db.Update(installation, tx)
		if err != nil {
			return err
		}

		// decrease the refers of the original plugin
		err = db.Run(
			db.WithTransactionContext(tx),
			db.Model(&models.Plugin{}),
			db.Equal("plugin_unique_identifier", original_plugin_unique_identifier.String()),
			db.Inc(map[string]int{"refers": -1}),
		)

		if err != nil {
			return err
		}

		// delete the original plugin if the refers is 0
		originalPlugin, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", original_plugin_unique_identifier.String()),
		)

		if err == nil && originalPlugin.Refers == 0 {
			err := db.Delete(&originalPlugin, tx)
			if err != nil {
				return err
			}
			response.IsOriginalPluginDeleted = true
			response.DeletedPlugin = &originalPlugin
		} else if err != nil {
			return err
		}

		// increase the refers of the new plugin
		err = db.Run(
			db.WithTransactionContext(tx),
			db.Model(&models.Plugin{}),
			db.Equal("plugin_unique_identifier", new_plugin_unique_identifier.String()),
			db.Inc(map[string]int{"refers": 1}),
		)

		if err != nil {
			return err
		}

		// update ai model installation
		if originalDeclaration.Model != nil {
			// delete the original ai model installation
			err := db.DeleteByCondition(&models.AIModelInstallation{
				PluginID: original_plugin_unique_identifier.PluginID(),
				TenantID: tenant_id,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.Model != nil {
			// create the new ai model installation
			modelInstallation := &models.AIModelInstallation{
				PluginUniqueIdentifier: new_plugin_unique_identifier.String(),
				TenantID:               tenant_id,
				Provider:               newDeclaration.Model.Provider,
				PluginID:               new_plugin_unique_identifier.PluginID(),
			}

			err := db.Create(modelInstallation, tx)
			if err != nil {
				return err
			}
		}

		// update tool installation
		if originalDeclaration.Tool != nil {
			// delete the original tool installation
			err := db.DeleteByCondition(&models.ToolInstallation{
				PluginID: original_plugin_unique_identifier.PluginID(),
				TenantID: tenant_id,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.Tool != nil {
			// create the new tool installation
			toolInstallation := &models.ToolInstallation{
				PluginUniqueIdentifier: new_plugin_unique_identifier.String(),
				TenantID:               tenant_id,
				Provider:               newDeclaration.Tool.Identity.Name,
				PluginID:               new_plugin_unique_identifier.PluginID(),
			}

			err := db.Create(toolInstallation, tx)
			if err != nil {
				return err
			}
		}

		// update agent installation
		if originalDeclaration.AgentStrategy != nil {
			// delete the original agent installation
			err := db.DeleteByCondition(&models.AgentStrategyInstallation{
				PluginID: original_plugin_unique_identifier.PluginID(),
				TenantID: tenant_id,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.AgentStrategy != nil {
			// create the new agent installation
			agentStrategyInstallation := &models.AgentStrategyInstallation{
				PluginUniqueIdentifier: new_plugin_unique_identifier.String(),
				TenantID:               tenant_id,
				Provider:               newDeclaration.AgentStrategy.Identity.Name,
				PluginID:               new_plugin_unique_identifier.PluginID(),
			}

			err := db.Create(agentStrategyInstallation, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &response, nil
}
