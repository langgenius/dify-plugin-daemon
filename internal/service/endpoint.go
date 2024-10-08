package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service/install_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func Endpoint(
	ctx *gin.Context,
	endpoint *models.Endpoint,
	plugin_installation *models.PluginInstallation,
	path string,
) {
	if !endpoint.Enabled {
		ctx.JSON(404, gin.H{"error": "plugin not found"})
		return
	}

	req := ctx.Request.Clone(context.Background())
	req.URL.Path = path

	var buffer bytes.Buffer
	err := req.Write(&buffer)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}

	identifier, err := plugin_entities.NewPluginUniqueIdentifier(plugin_installation.PluginUniqueIdentifier)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid plugin identifier, " + err.Error()})
		return
	}

	// fetch plugin
	manager := plugin_manager.Manager()
	runtime := manager.Get(identifier)
	if runtime == nil {
		ctx.JSON(404, gin.H{"error": "plugin not found"})
		return
	}

	// fetch endpoint declaration
	endpoint_declaration := runtime.Configuration().Endpoint
	if endpoint_declaration == nil {
		ctx.JSON(404, gin.H{"error": "endpoint declaration not found"})
		return
	}

	// decrypt settings
	settings, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: endpoint.TenantID,
			UserId:   "",
			Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
		},
		InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
			Opt:       dify_invocation.ENCRYPT_OPT_DECRYPT,
			Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
			Identity:  endpoint.ID,
			Data:      endpoint.Settings,
			Config:    endpoint_declaration.Settings,
		},
	})

	if err != nil {
		ctx.JSON(500, gin.H{"error": "failed to decrypt data"})
		return
	}

	session := session_manager.NewSession(
		session_manager.NewSessionPayload{
			TenantID:               endpoint.TenantID,
			UserID:                 "",
			PluginUniqueIdentifier: identifier,
			ClusterID:              ctx.GetString("cluster_id"),
			InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_ENDPOINT,
			Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT,
			Declaration:            runtime.Configuration(),
			BackwardsInvocation:    manager.BackwardsInvocation(),
			IgnoreCache:            false,
		},
	)
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

	session.BindRuntime(runtime)

	status_code, headers, response, err := plugin_daemon.InvokeEndpoint(
		session, &requests.RequestInvokeEndpoint{
			RawHttpRequest: hex.EncodeToString(buffer.Bytes()),
			Settings:       settings,
		},
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer response.Close()

	done := make(chan bool)
	closed := new(int32)

	ctx.Status(status_code)
	for k, v := range *headers {
		if len(v) > 0 {
			ctx.Writer.Header().Set(k, v[0])
		}
	}

	close := func() {
		if atomic.CompareAndSwapInt32(closed, 0, 1) {
			close(done)
		}
	}
	defer close()

	routine.Submit(func() {
		defer close()
		for response.Next() {
			chunk, err := response.Read()
			if err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			ctx.Writer.Write(chunk)
			ctx.Writer.Flush()
		}
	})

	select {
	case <-ctx.Writer.CloseNotify():
	case <-done:
	case <-time.After(240 * time.Second):
		ctx.JSON(500, gin.H{"error": "killed by timeout"})
	}
}

func EnableEndpoint(endpoint_id string, tenant_id string) *entities.Response {
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("id", endpoint_id),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, "Endpoint not found")
	}

	endpoint.Enabled = true

	if err := install_service.EnabledEndpoint(&endpoint); err != nil {
		return entities.NewErrorResponse(-500, "Failed to enable endpoint")
	}

	return entities.NewSuccessResponse(true)
}

func DisableEndpoint(endpoint_id string, tenant_id string) *entities.Response {
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("id", endpoint_id),
		db.Equal("tenant_id", tenant_id),
	)
	if err != nil {
		return entities.NewErrorResponse(-404, "Endpoint not found")
	}

	endpoint.Enabled = false

	if err := install_service.DisabledEndpoint(&endpoint); err != nil {
		return entities.NewErrorResponse(-500, "Failed to disable endpoint")
	}

	return entities.NewSuccessResponse(true)
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

	manager := plugin_manager.Manager()
	if manager == nil {
		return entities.NewErrorResponse(-500, "failed to get plugin manager")
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

		plugin_unique_identifier, err := plugin_entities.NewPluginUniqueIdentifier(
			plugin_installation.PluginUniqueIdentifier,
		)
		if err != nil {
			return entities.NewErrorResponse(-500, fmt.Sprintf("failed to parse plugin unique identifier: %v", err))
		}

		plugin_declaration, err := helper.CombinedGetPluginDeclaration(plugin_unique_identifier)
		if err != nil {
			return entities.NewErrorResponse(-500, fmt.Sprintf("failed to get plugin declaration: %v", err))
		}

		if plugin_declaration.Endpoint == nil {
			return entities.NewErrorResponse(-404, "plugin does not have an endpoint")
		}

		decrypted_settings, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   "",
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_DECRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  endpoint.ID,
				Data:      endpoint.Settings,
				Config:    plugin_declaration.Endpoint.Settings,
			},
		})
		if err != nil {
			return entities.NewErrorResponse(-500, fmt.Sprintf("failed to decrypt settings: %v", err))
		}

		endpoint.Settings = decrypted_settings
		endpoint.Declaration = plugin_declaration.Endpoint

		endpoints[i] = endpoint
	}

	return entities.NewSuccessResponse(endpoints)
}

func ListPluginEndpoints(tenant_id string, plugin_unique_identifier plugin_entities.PluginUniqueIdentifier, page int, page_size int) *entities.Response {
	endpoints, err := db.GetAll[models.Endpoint](
		db.Equal("plugin_id", plugin_unique_identifier.PluginID()),
		db.Equal("tenant_id", tenant_id),
		db.OrderBy("created_at", true),
		db.Page(page, page_size),
	)
	if err != nil {
		return entities.NewErrorResponse(-500, fmt.Sprintf("failed to list endpoints: %v", err))
	}

	manager := plugin_manager.Manager()
	if manager == nil {
		return entities.NewErrorResponse(-500, "failed to get plugin manager")
	}

	// decrypt settings
	for i, endpoint := range endpoints {
		plugin_declaration, err := helper.CombinedGetPluginDeclaration(plugin_unique_identifier)
		if err != nil {
			return entities.NewErrorResponse(-500, fmt.Sprintf("failed to get plugin declaration: %v", err))
		}

		decrypted_settings, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
			BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
				TenantId: tenant_id,
				UserId:   "",
				Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
			},
			InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
				Opt:       dify_invocation.ENCRYPT_OPT_DECRYPT,
				Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
				Identity:  endpoint.ID,
				Data:      endpoint.Settings,
				Config:    plugin_declaration.Endpoint.Settings,
			},
		})
		if err != nil {
			return entities.NewErrorResponse(-500, fmt.Sprintf("failed to decrypt settings: %v", err))
		}

		endpoint.Settings = decrypted_settings
		endpoint.Declaration = plugin_declaration.Endpoint

		endpoints[i] = endpoint
	}

	return entities.NewSuccessResponse(endpoints)
}
