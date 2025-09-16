package service

import (
	"errors"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service/install_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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
		return exception.ErrPluginNotFound().ToResponse()
	}

	// try get plugin
	pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
		pluginUniqueIdentifier,
		plugin_entities.PluginRuntimeType(installation.RuntimeType),
	)

	if err != nil {
		return exception.ErrPluginNotFound().ToResponse()
	}

	if !pluginDeclaration.Resource.Permission.AllowRegisterEndpoint() {
		return exception.PermissionDeniedError("permission denied, you need to enable endpoint access in plugin manifest").ToResponse()
	}

	if pluginDeclaration.Endpoint == nil {
		return exception.BadRequestError(errors.New("plugin does not have an endpoint")).ToResponse()
	}

	// check settings
	if err := plugin_entities.ValidateProviderConfigs(settings, pluginDeclaration.Endpoint.Settings); err != nil {
		return exception.BadRequestError(fmt.Errorf("failed to validate settings: %v", err)).ToResponse()
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
		return exception.InternalServerError(fmt.Errorf("failed to setup endpoint: %v", err)).ToResponse()
	}

	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(errors.New("failed to get plugin manager")).ToResponse()
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
		return exception.InternalServerError(fmt.Errorf("failed to encrypt settings: %v", err)).ToResponse()
	}

	if err := install_service.UpdateEndpoint(endpoint, name, encryptedSettings); err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to update endpoint: %v", err)).ToResponse()
	}

	return entities.NewSuccessResponse(true)
}

func RemoveEndpoint(endpoint_id string, tenant_id string) *entities.Response {
	endpoint, err := install_service.UninstallEndpoint(endpoint_id, tenant_id)
	if err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to remove endpoint: %v", err)).ToResponse()
	}

	// invalidate endpoint cache
	endpointCacheKey := helper.EndpointCacheKey(endpoint.HookID)
	_, _ = cache.AutoDelete[models.Endpoint](endpointCacheKey)

	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(errors.New("failed to get plugin manager")).ToResponse()
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
		return exception.InternalServerError(fmt.Errorf("failed to clear credentials cache: %v", err)).ToResponse()
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
		return exception.NotFoundError(fmt.Errorf("failed to find endpoint: %v", err)).ToResponse()
	}

	// get plugin installation
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", endpoint.PluginID),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return exception.NotFoundError(fmt.Errorf("failed to find plugin installation: %v", err)).ToResponse()
	}

	pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
		installation.PluginUniqueIdentifier,
	)
	if err != nil {
		return exception.UniqueIdentifierError(fmt.Errorf("failed to parse plugin unique identifier: %v", err)).ToResponse()
	}

	// get plugin
	pluginDeclaration, err := helper.CombinedGetPluginDeclaration(
		pluginUniqueIdentifier,
		plugin_entities.PluginRuntimeType(installation.RuntimeType),
	)
	if err != nil {
		return exception.ErrPluginNotFound().ToResponse()
	}

	if pluginDeclaration.Endpoint == nil {
		return exception.BadRequestError(errors.New("plugin does not have an endpoint")).ToResponse()
	}

	// decrypt original settings
	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(errors.New("failed to get plugin manager")).ToResponse()
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
		return exception.InternalServerError(fmt.Errorf("failed to decrypt settings: %v", err)).ToResponse()
	}

	maskedSettings := encryption.MaskConfigCredentials(originalSettings, pluginDeclaration.Endpoint.Settings)

	// check if settings is changed, replace the value is the same as masked_settings
	for settingName, value := range settings {
		// skip it if the value is not secret-input
		found := false
		for _, config := range pluginDeclaration.Endpoint.Settings {
			if config.Name == settingName && config.Type == plugin_entities.CONFIG_TYPE_SECRET_INPUT {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		if maskedSettings[settingName] == value {
			settings[settingName] = originalSettings[settingName]
		}
	}

	// check settings
	if err := plugin_entities.ValidateProviderConfigs(settings, pluginDeclaration.Endpoint.Settings); err != nil {
		return exception.BadRequestError(fmt.Errorf("failed to validate settings: %v", err)).ToResponse()
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
		return exception.InternalServerError(fmt.Errorf("failed to encrypt settings: %v", err)).ToResponse()
	}

	// update endpoint
	if err := install_service.UpdateEndpoint(&endpoint, name, encryptedSettings); err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to update endpoint: %v", err)).ToResponse()
	}

	// invalidate endpoint cache
	endpointCacheKey := helper.EndpointCacheKey(endpoint.HookID)
	_, _ = cache.AutoDelete[models.Endpoint](endpointCacheKey)

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
		return exception.InternalServerError(fmt.Errorf("failed to clear credentials cache: %v", err)).ToResponse()
	}

	return entities.NewSuccessResponse(true)
}
