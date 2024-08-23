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
	configuration map[string]any,
) (string, error) {
	identity, err := runtime.Identity()
	if err != nil {
		return "", err
	}

	plugin := &models.Plugin{
		PluginID:     identity,
		Refers:       0,
		Checksum:     runtime.Checksum(),
		InstallType:  runtime.Type(),
		ManifestType: runtime.Configuration().Type,
	}

	plugin, installation, err := curd.CreatePlugin(tenant_id, user_id, plugin, configuration)
	if err != nil {
		return "", err
	}

	// check if there is a webhook for the plugin
	if runtime.Configuration().Resource.Permission.AllowRegistryWebhook() {
		_, err := InstallWebhook(plugin.PluginID, installation.ID, tenant_id, user_id)
		if err != nil {
			return "", err
		}
	}

	return installation.ID, nil
}

func UninstallPlugin(tenant_id string, installation_id string, runtime plugin_entities.PluginRuntimeInterface) error {
	identity, err := runtime.Identity()
	if err != nil {
		return err
	}

	// delete the plugin from db
	resp, err := curd.DeletePlugin(tenant_id, identity, installation_id)
	if err != nil {
		return err
	}

	// delete the webhook from db
	if runtime.Configuration().Resource.Permission.AllowRegistryWebhook() {
		// get the webhook from db
		webhook, err := GetWebhook(tenant_id, identity, resp.Installation.ID)
		if err != nil && err != db.ErrDatabaseNotFound {
			return err
		} else if err == db.ErrDatabaseNotFound {
			return nil
		}

		err = UninstallWebhook(webhook)
		if err != nil {
			return err
		}
	}

	return nil
}

// installs a plugin to db,
// returns the webhook id
func InstallWebhook(plugin_id string, installation_id string, tenant_id string, user_id string) (string, error) {
	installation := &models.Webhook{
		HookID:               strings.RandomString(64),
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

func GetWebhook(tenant_id string, plugin_id string, installation_id string) (*models.Webhook, error) {
	webhook, err := db.GetOne[models.Webhook](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_id", plugin_id),
		db.Equal("plugin_installation_id", installation_id),
	)

	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

// uninstalls a plugin from db
func UninstallWebhook(webhook *models.Webhook) error {
	return db.Delete(webhook)
}
