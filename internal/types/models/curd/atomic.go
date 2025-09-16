package curd

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"gorm.io/gorm"
)

func EnsureGlobalReferenceIfRequired(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	tenantId string,
	installType plugin_entities.PluginRuntimeType,
	declaration *plugin_entities.PluginDeclaration,
	source string,
	meta map[string]any,
) error {
	if !allowOrphans || tenantId == constants.GlobalTenantId {
		return nil
	}
	_, _, err := InstallPlugin(
		constants.GlobalTenantId,
		pluginUniqueIdentifier,
		installType,
		declaration,
		source,
		meta,
	)
	if err != nil && err != ErrPluginAlreadyInstalled {
		return err
	}
	return nil
}

// Create plugin for a tenant, create plugin if it has never been created before
// and install it to the tenant, return the plugin and the installation
// if the plugin has been created before, return the plugin which has been created before
func InstallPlugin(
	tenantId string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	installType plugin_entities.PluginRuntimeType,
	declaration *plugin_entities.PluginDeclaration,
	source string,
	meta map[string]any,
) (*models.Plugin, *models.PluginInstallation, error) {

	var pluginToBeReturns *models.Plugin
	var installationToBeReturns *models.PluginInstallation

	// check if already installed
	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", pluginUniqueIdentifier.PluginID()),
		db.Equal("tenant_id", tenantId),
	)

	if err == nil {
		return nil, nil, ErrPluginAlreadyInstalled
	}

	err = db.WithTransaction(func(tx *gorm.DB) error {
		p, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
			db.Equal("plugin_id", pluginUniqueIdentifier.PluginID()),
			db.Equal("install_type", string(installType)),
			db.WLock(),
		)

		if err == db.ErrDatabaseNotFound {
			plugin := &models.Plugin{
				PluginID:               pluginUniqueIdentifier.PluginID(),
				PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
				InstallType:            installType,
				Refers:                 1,
				Source:                 source,
			}

			if installType == plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
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
				RuntimeType: string(installType),
				TenantID:    tenantId,
			},
			tx,
		); err != nil {
			return err
		}

		installation := &models.PluginInstallation{
			PluginID:               pluginToBeReturns.PluginID,
			PluginUniqueIdentifier: pluginToBeReturns.PluginUniqueIdentifier,
			TenantID:               tenantId,
			RuntimeType:            string(installType),
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
				TenantID:               tenantId,
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
				TenantID:               tenantId,
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
				TenantID:               tenantId,
				Provider:               declaration.Model.Provider,
			}

			err := db.Create(modelInstallation, tx)
			if err != nil {
				return err
			}
		}

		// create datasource installation
		if declaration.Datasource != nil {
			datasourceInstallation := &models.DatasourceInstallation{
				PluginID:               pluginToBeReturns.PluginID,
				PluginUniqueIdentifier: pluginToBeReturns.PluginUniqueIdentifier,
				TenantID:               tenantId,
				Provider:               declaration.Datasource.Identity.Name,
			}

			err := db.Create(datasourceInstallation, tx)
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
	tenantId string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	installationId string,
	declaration *plugin_entities.PluginDeclaration,
) (*DeletePluginResponse, error) {
	var pluginToBeReturns *models.Plugin
	var installationToBeReturns *models.PluginInstallation

	_, err := db.GetOne[models.PluginInstallation](
		db.Equal("id", installationId),
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
		db.Equal("tenant_id", tenantId),
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
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
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
			db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
			db.Equal("tenant_id", tenantId),
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
				TenantID: tenantId,
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
				TenantID: tenantId,
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
				TenantID: tenantId,
			}

			err := db.DeleteByCondition(&modelInstallation, tx)
			if err != nil {
				return err
			}
		}

		// delete datasource installation
		if declaration.Datasource != nil {
			datasourceInstallation := &models.DatasourceInstallation{
				PluginID: pluginToBeReturns.PluginID,
				TenantID: tenantId,
			}

			err := db.DeleteByCondition(&datasourceInstallation, tx)
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
	tenantId string,
	originalPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	newPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	originalDeclaration *plugin_entities.PluginDeclaration,
	newDeclaration *plugin_entities.PluginDeclaration,
	installType plugin_entities.PluginRuntimeType,
	source string,
	meta map[string]any,
) (*UpgradePluginResponse, error) {
	var response UpgradePluginResponse

	err := db.WithTransaction(func(tx *gorm.DB) error {
		installation, err := db.GetOne[models.PluginInstallation](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", originalPluginUniqueIdentifier.String()),
			db.Equal("tenant_id", tenantId),
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
			db.Equal("plugin_unique_identifier", newPluginUniqueIdentifier.String()),
		)

		if err == db.ErrDatabaseNotFound {
			// create new plugin
			plugin = models.Plugin{
				PluginID:               newPluginUniqueIdentifier.PluginID(),
				PluginUniqueIdentifier: newPluginUniqueIdentifier.String(),
				InstallType:            installType,
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
		installation.PluginUniqueIdentifier = newPluginUniqueIdentifier.String()
		installation.Meta = meta
		err = db.Update(installation, tx)
		if err != nil {
			return err
		}

		// decrease the refers of the original plugin
		err = db.Run(
			db.WithTransactionContext(tx),
			db.Model(&models.Plugin{}),
			db.Equal("plugin_unique_identifier", originalPluginUniqueIdentifier.String()),
			db.Inc(map[string]int{"refers": -1}),
		)

		if err != nil {
			return err
		}

		// delete the original plugin if the refers is 0
		originalPlugin, err := db.GetOne[models.Plugin](
			db.WithTransactionContext(tx),
			db.Equal("plugin_unique_identifier", originalPluginUniqueIdentifier.String()),
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
			db.Equal("plugin_unique_identifier", newPluginUniqueIdentifier.String()),
			db.Inc(map[string]int{"refers": 1}),
		)

		if err != nil {
			return err
		}

		// update ai model installation
		if originalDeclaration.Model != nil {
			// delete the original ai model installation
			err := db.DeleteByCondition(&models.AIModelInstallation{
				PluginID: originalPluginUniqueIdentifier.PluginID(),
				TenantID: tenantId,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.Model != nil {
			// create the new ai model installation
			modelInstallation := &models.AIModelInstallation{
				PluginUniqueIdentifier: newPluginUniqueIdentifier.String(),
				TenantID:               tenantId,
				Provider:               newDeclaration.Model.Provider,
				PluginID:               newPluginUniqueIdentifier.PluginID(),
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
				PluginID: originalPluginUniqueIdentifier.PluginID(),
				TenantID: tenantId,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.Tool != nil {
			// create the new tool installation
			toolInstallation := &models.ToolInstallation{
				PluginUniqueIdentifier: newPluginUniqueIdentifier.String(),
				TenantID:               tenantId,
				Provider:               newDeclaration.Tool.Identity.Name,
				PluginID:               newPluginUniqueIdentifier.PluginID(),
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
				PluginID: originalPluginUniqueIdentifier.PluginID(),
				TenantID: tenantId,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.AgentStrategy != nil {
			// create the new agent installation
			agentStrategyInstallation := &models.AgentStrategyInstallation{
				PluginUniqueIdentifier: newPluginUniqueIdentifier.String(),
				TenantID:               tenantId,
				Provider:               newDeclaration.AgentStrategy.Identity.Name,
				PluginID:               newPluginUniqueIdentifier.PluginID(),
			}

			err := db.Create(agentStrategyInstallation, tx)
			if err != nil {
				return err
			}
		}

		// update datasource installation
		if originalDeclaration.Datasource != nil {
			// delete the original datasource installation
			err := db.DeleteByCondition(&models.DatasourceInstallation{
				PluginID: originalPluginUniqueIdentifier.PluginID(),
				TenantID: tenantId,
			}, tx)

			if err != nil {
				return err
			}
		}

		if newDeclaration.Datasource != nil {
			// create the new datasource installation
			datasourceInstallation := &models.DatasourceInstallation{
				PluginUniqueIdentifier: newPluginUniqueIdentifier.String(),
				TenantID:               tenantId,
				Provider:               newDeclaration.Datasource.Identity.Name,
				PluginID:               newPluginUniqueIdentifier.PluginID(),
			}

			err := db.Create(datasourceInstallation, tx)
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
