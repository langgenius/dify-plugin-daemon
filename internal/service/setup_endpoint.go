package service

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service/install_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
)

func SetupEndpoint(
	tenant_id string,
	user_id string,
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	name string,
	settings map[string]any,
) *entities.Response {
	// try find plugin installation
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin installation: %v", err))
	}

	// try get plugin
	pluginDeclaration, err := helper.CombinedGetPluginDeclaration(pluginUniqueIdentifier)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin: %v", err))
	}

	if !pluginDeclaration.Resource.Permission.AllowRegisterEndpoint() {
		return entities.NewErrorResponse(-403, "permission denied")
	}

	if pluginDeclaration.Endpoint == nil {
		return entities.NewErrorResponse(-404, "plugin does not have an endpoint")
	}

	// check settings
	if err := plugin_entities.ValidateProviderConfigs(settings, pluginDeclaration.Endpoint.Settings); err != nil {
		return entities.NewErrorResponse(-400, fmt.Sprintf("failed to validate settings: %v", err))
	}

	endpoint, err := install_service.InstallEndpoint(
		pluginUniqueIdentifier,
		installation.ID,
		tenant_id,
		user_id,
		name,
		map[string]any{},
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to setup endpoint: %v", err))
	}

	manager := plugin_manager.Manager()
	if manager == nil {
		return entities.NewErrorResponse(-500, "failed to get plugin manager")
	}

	// encrypt settings
	encryptedSettings, err := manager.BackwardsInvocation().InvokeEncrypt(
		&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   user_id,
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_ENCRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  endpoint.ID,
				Data:      settings,
				Config:    pluginDeclaration.Endpoint.Settings,
			},
		},
	)

	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to encrypt settings: %v", err))
	}

	if err := install_service.UpdateEndpoint(endpoint, name, encryptedSettings); err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to update endpoint: %v", err))
	}

	return entities.NewSuccessResponse(true)
}

func RemoveEndpoint(endpoint_id string, tenant_id string) *entities.Response {
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("id", endpoint_id),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find endpoint: %v", err))
	}

	err = install_service.UninstallEndpoint(&endpoint)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to remove endpoint: %v", err))
	}

	manager := plugin_manager.Manager()
	if manager == nil {
		return entities.NewErrorResponse(-500, "failed to get plugin manager")
	}

	// clear credentials cache
	if _, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: tenant_id,
			UserId:   "",
			Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
		},
		InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
			Opt:       dify_invocation.ENCRYPT_OPT_CLEAR,
			Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
			Identity:  endpoint.ID,
		},
	}); err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to clear credentials cache: %v", err))
	}

	return entities.NewSuccessResponse(true)
}

func UpdateEndpoint(endpoint_id string, tenant_id string, user_id string, name string, settings map[string]any) *entities.Response {
	// get endpoint
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("id", endpoint_id),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find endpoint: %v", err))
	}

	// get plugin installation
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", endpoint.PluginID),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin installation: %v", err))
	}

	pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
		installation.PluginUniqueIdentifier,
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to parse plugin unique identifier: %v", err))
	}

	// get plugin
	pluginDeclaration, err := helper.CombinedGetPluginDeclaration(pluginUniqueIdentifier)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin: %v", err))
	}

	if pluginDeclaration.Endpoint == nil {
		return entities.NewErrorResponse(-404, "plugin does not have an endpoint")
	}

	// decrypt original settings
	manager := plugin_manager.Manager()
	if manager == nil {
		return entities.NewErrorResponse(-500, "failed to get plugin manager")
	}

	originalSettings, err := manager.BackwardsInvocation().InvokeEncrypt(
		&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   user_id,
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_DECRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  endpoint.ID,
				Data:      endpoint.Settings,
				Config:    pluginDeclaration.Endpoint.Settings,
			},
		},
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to decrypt settings: %v", err))
	}

	maskedSettings := encryption.MaskConfigCredentials(originalSettings, pluginDeclaration.Endpoint.Settings)

	// check if settings is changed, replace the value is the same as masked_settings
	for setting_name, value := range settings {
		if maskedSettings[setting_name] == value {
			settings[setting_name] = originalSettings[setting_name]
		}
	}

	// check settings
	if err := plugin_entities.ValidateProviderConfigs(settings, pluginDeclaration.Endpoint.Settings); err != nil {
		return entities.NewErrorResponse(-400, fmt.Sprintf("failed to validate settings: %v", err))
	}

	// encrypt settings
	encryptedSettings, err := manager.BackwardsInvocation().InvokeEncrypt(
		&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   user_id,
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_ENCRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  endpoint.ID,
				Data:      settings,
				Config:    pluginDeclaration.Endpoint.Settings,
			},
		},
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to encrypt settings: %v", err))
	}

	// update endpoint
	if err := install_service.UpdateEndpoint(&endpoint, name, encryptedSettings); err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to update endpoint: %v", err))
	}

	// clear credentials cache
	if _, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: tenant_id,
			UserId:   user_id,
			Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
		},
		InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
			Opt:       dify_invocation.ENCRYPT_OPT_CLEAR,
			Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
			Identity:  endpoint.ID,
			Data:      settings,
			Config:    pluginDeclaration.Endpoint.Settings,
		},
	}); err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to clear credentials cache: %v", err))
	}

	return entities.NewSuccessResponse(true)
}
