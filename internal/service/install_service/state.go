package install_service

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

func InstallPlugin(
	tenant_id string,
	user_id string,
	runtime entities.PluginRuntimeInterface,
	configuration map[string]any,
) error {
	identity, err := runtime.Identity()
	if err != nil {
		return err
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
		return err
	}

	// check if there is a webhook for the plugin
	if runtime.Configuration().Resource.Permission.AllowRegistryWebhook() {
		_, err := InstallWebhook(plugin.PluginID, installation.ID, tenant_id, user_id)
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

// uninstalls a plugin from db
func UninstallWebhook(webhook *models.Webhook) error {
	return db.Delete(webhook)
}
