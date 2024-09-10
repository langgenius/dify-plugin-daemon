package service

import (
	"bytes"
	"context"
	"encoding/hex"
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
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func Endpoint(
	ctx *gin.Context,
	endpoint *models.Endpoint,
	plugin_installation *models.PluginInstallation,
	path string,
) {
	req := ctx.Request.Clone(context.Background())
	req.URL.Path = path

	var buffer bytes.Buffer
	err := req.Write(&buffer)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}

	// fetch plugin
	manager := plugin_manager.Manager()
	runtime := manager.Get(
		plugin_entities.PluginUniqueIdentifier(plugin_installation.PluginUniqueIdentifier),
	)
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
	settings, err := dify_invocation.InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: endpoint.TenantID,
			UserId:   "",
			Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
		},
		InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
			Opt:       dify_invocation.ENCRYPT_OPT_DECRYPT,
			Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
			Identity:  endpoint.ID,
			Data:      endpoint.GetSettings(),
			Config:    endpoint_declaration.Settings,
		},
	})

	if err != nil {
		ctx.JSON(500, gin.H{"error": "failed to decrypt data"})
		return
	}

	session := session_manager.NewSession(
		endpoint.TenantID,
		"",
		plugin_entities.PluginUniqueIdentifier(plugin_installation.PluginUniqueIdentifier),
		ctx.GetString("cluster_id"),
		access_types.PLUGIN_ACCESS_TYPE_ENDPOINT,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT,
		runtime.Configuration(),
	)
	defer session.Close()

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
	case <-time.After(30 * time.Second):
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

	return entities.NewSuccessResponse("success")
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

	return entities.NewSuccessResponse("success")
}
