package service

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service/install_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
)

func SetupEndpoint(
	tenant_id string,
	user_id string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	settings map[string]any,
) *entities.Response {
	// try find plugin installation
	installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("tenant_id", tenant_id),
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin installation: %v", err))
	}

	// try get plugin
	plugin, err := db.GetOne[models.Plugin](
		db.Equal("plugin_unique_identifier", plugin_unique_identifier.String()),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin: %v", err))
	}

	declaration, err := plugin.GetDeclaration()
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to get plugin declaration: %v", err))
	}

	if !declaration.Resource.Permission.AllowRegisterEndpoint() {
		return entities.NewErrorResponse(-403, "permission denied")
	}

	if declaration.Endpoint == nil {
		return entities.NewErrorResponse(-404, "plugin does not have an endpoint")
	}

	// encrypt settings
	encrypted_settings, err := dify_invocation.InvokeEncrypt(
		&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   user_id,
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_ENCRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  installation.ID,
				Data:      settings,
				Config:    declaration.Endpoint.Settings,
			},
		},
	)

	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to encrypt settings: %v", err))
	}

	_, err = install_service.InstallEndpoint(
		plugin_unique_identifier,
		installation.ID,
		tenant_id,
		user_id,
		encrypted_settings,
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to setup endpoint: %v", err))
	}

	return entities.NewSuccessResponse(nil)
}

func ListEndpoints(tenant_id string, page int, page_size int) *entities.Response {
	endpoints, err := db.GetAll[models.Endpoint](
		db.Equal("tenant_id", tenant_id),
		db.OrderBy("created_at", true),
		db.Page(page, page_size),
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to list endpoints: %v", err))
	}

	// decrypt settings
	for i, endpoint := range endpoints {
		plugin_installation, err := db.GetOne[models.PluginInstallation](
			db.Equal("plugin_id", endpoint.PluginID),
			db.Equal("tenant_id", tenant_id),
		)
		if err != nil {
			return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin installation: %v", err))
		}

		plugin, err := db.GetOne[models.Plugin](
			db.Equal("plugin_unique_identifier", plugin_installation.PluginUniqueIdentifier),
		)
		if err != nil {
			return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find plugin: %v", err))
		}

		plugin_declaration, err := plugin.GetDeclaration()
		if err != nil {
			return entities.NewErrorResponse(-404, fmt.Sprintf("failed to get plugin declaration: %v", err))
		}

		if plugin_declaration.Endpoint == nil {
			return entities.NewErrorResponse(-404, "plugin does not have an endpoint")
		}

		decrypted_settings, err := dify_invocation.InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   "",
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_DECRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  endpoint.ID,
				Data:      endpoint.GetSettings(),
				Config:    plugin_declaration.Endpoint.Settings,
			},
		})
		if err != nil {
			return entities.NewErrorResponse(-500, fmt.Sprintf("failed to decrypt settings: %v", err))
		}

		endpoint.SetSettings(decrypted_settings)

		endpoints[i] = endpoint
	}

	return entities.NewSuccessResponse(endpoints)
}

func RemoveEndpoint(endpoint_id string, tenant_id string) *entities.Response {
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("endpoint_id", endpoint_id),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, fmt.Sprintf("failed to find endpoint: %v", err))
	}

	err = install_service.UninstallEndpoint(&endpoint)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to remove endpoint: %v", err))
	}

	return entities.NewSuccessResponse(nil)
}
