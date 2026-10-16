package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service/install_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/endpoint_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func copyRequest(req *http.Request, hookId string, path string) (*bytes.Buffer, error) {
	newReq := req.Clone(context.Background())
	// get query params
	queryParams := req.URL.Query()

	// replace path with endpoint path
	newReq.URL.Path = path
	// set query params
	newReq.URL.RawQuery = queryParams.Encode()

	// read request body until complete, max 10MB
	body, err := io.ReadAll(io.LimitReader(req.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}

	// replace with a new reader
	newReq.Body = io.NopCloser(bytes.NewReader(body))
	newReq.ContentLength = int64(len(body))
	newReq.TransferEncoding = nil

	// remove ip traces for security
	newReq.Header.Del("X-Forwarded-For")
	newReq.Header.Del("X-Real-IP")
	newReq.Header.Del("X-Forwarded")
	newReq.Header.Del("X-Original-Forwarded-For")
	newReq.Header.Del("X-Original-Url")
	newReq.Header.Del("X-Original-Host")

	// check if X-Original-Host is set
	if originalHost := req.Header.Get(endpoint_entities.HeaderXOriginalHost); originalHost != "" {
		// replace host with original host
		newReq.Host = originalHost
	}

	// setup hook id to request
	newReq.Header.Set("Dify-Hook-Id", hookId)
	// check if Dify-Hook-Url is set
	if url := req.Header.Get(endpoint_entities.HeaderDifyHookURL); url == "" {
		scheme := "http"
		if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		newReq.Header.Set(
			endpoint_entities.HeaderDifyHookURL,
			fmt.Sprintf("%s://%s/e/%s%s", scheme, newReq.Host, hookId, path),
		)
	}

	var buffer bytes.Buffer
	err = newReq.Write(&buffer)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func Endpoint(
	ctx *gin.Context,
	endpoint *models.Endpoint,
	pluginInstallation *models.PluginInstallation,
	maxExecutionTime time.Duration,
	path string,
) {
	if !endpoint.Enabled {
		ctx.JSON(404, exception.NotFoundError(errors.New("endpoint not found")).ToResponse())
		return
	}

	buffer, err := copyRequest(ctx.Request, endpoint.HookID, path)
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}

	identifier, err := plugin_entities.NewPluginUniqueIdentifier(pluginInstallation.PluginUniqueIdentifier)
	if err != nil {
		ctx.JSON(400, exception.UniqueIdentifierError(err).ToResponse())
		return
	}

	// fetch plugin
	manager := plugin_manager.Manager()
	runtime, err := manager.GetPluginRuntime(identifier)
	if err != nil {
		ctx.JSON(404, exception.ErrPluginNotFound().ToResponse())
		return
	}

	// fetch endpoint declaration
	endpointDeclaration := runtime.Configuration().Endpoint
	if endpointDeclaration == nil {
		ctx.JSON(404, exception.ErrPluginNotFound().ToResponse())
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
			Config:    endpointDeclaration.Settings,
		},
	})

	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}

	sessionCtx, cancelSessionCtx := endpointSessionRequestContext(ctx.Request.Context(), maxExecutionTime)
	defer cancelSessionCtx()

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
			EndpointID:             &endpoint.ID,
			RequestContext:         sessionCtx,
		},
	)
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

	session.BindRuntime(runtime)

	statusCode, headers, response, err := io_tunnel.InvokeEndpoint(
		session, &requests.RequestInvokeEndpoint{
			RawHttpRequest: hex.EncodeToString(buffer.Bytes()),
			Settings:       settings,
		},
	)
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	writeEndpointResponse(ctx, sessionCtx, statusCode, *headers, response)
}

func writeEndpointResponse(
	ctx *gin.Context,
	sessionCtx context.Context,
	statusCode int,
	headers http.Header,
	response *stream.Stream[[]byte],
) {
	defer response.Close()
	type chunk struct {
		data []byte
		err  error
	}
	chunks := make(chan chunk)
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		defer close(chunks)
		for response.Next() {
			data, err := response.Read()
			select {
			case chunks <- chunk{data, err}:
			case <-stop:
				return
			}
			if err != nil {
				return
			}
		}
	}()

	connected := true
	clientDone := ctx.Request.Context().Done()
	fail := func(err error) {
		failure := exception.InternalServerError(err).ToResponse()
		if !connected {
			return
		}
		if ctx.Writer.Written() {
			panic(http.ErrAbortHandler)
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, failure)
	}
	setHeaders := func() {
		if !ctx.Writer.Written() {
			ctx.Status(statusCode)
			for k, v := range headers {
				ctx.Writer.Header()[k] = v
			}
		}
	}

	for {
		select {
		case <-sessionCtx.Done():
			fail(sessionCtx.Err())
			return
		case <-clientDone:
			// Keep the plugin session alive for backwards invocations after disconnect (#808).
			connected = false
			clientDone = nil
		case chunk, ok := <-chunks:
			if err := sessionCtx.Err(); err != nil {
				fail(err)
				return
			}
			if chunk.err != nil {
				fail(chunk.err)
				return
			}
			if !ok {
				if connected {
					setHeaders()
				}
				return
			}
			if !connected || ctx.Request.Context().Err() != nil {
				connected = false
				clientDone = nil
				continue
			}
			if !ctx.Writer.Written() {
				if deadline, ok := sessionCtx.Deadline(); ok {
					err := http.NewResponseController(ctx.Writer).SetWriteDeadline(deadline)
					if err != nil && !errors.Is(err, http.ErrNotSupported) {
						fail(err)
						return
					}
				}
				setHeaders()
			}
			if _, err := ctx.Writer.Write(chunk.data); err != nil {
				log.Error("failed to write endpoint response", "error", err)
				connected = false
				clientDone = nil
				continue
			}
			ctx.Writer.Flush()
		}
	}
}

// endpointSessionRequestContext preserves request trace metadata while decoupling
// plugin session lifetime from HTTP request cancellation (see issue #808).
func endpointSessionRequestContext(
	httpCtx context.Context,
	maxExecutionTime time.Duration,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(httpCtx), maxExecutionTime)
}

func EnableEndpoint(endpointID string, tenantID string) *entities.Response {
	endpoint, err := install_service.EnabledEndpoint(endpointID, tenantID)
	if err != nil {
		return handleEndpointStateError(err, "enable")
	}

	invalidateEndpointCache(endpoint.HookID)

	return entities.NewSuccessResponse(true)
}

func DisableEndpoint(endpointID string, tenantID string) *entities.Response {
	endpoint, err := install_service.DisabledEndpoint(endpointID, tenantID)
	if err != nil {
		return handleEndpointStateError(err, "disable")
	}

	invalidateEndpointCache(endpoint.HookID)

	return entities.NewSuccessResponse(true)
}

func handleEndpointStateError(err error, action string) *entities.Response {
	if err == nil {
		return nil
	}
	if errors.Is(err, db.ErrDatabaseNotFound) {
		return exception.NotFoundError(errors.New("endpoint not found")).ToResponse()
	}
	return exception.InternalServerError(fmt.Errorf("failed to %s endpoint: %w", action, err)).ToResponse()
}

func invalidateEndpointCache(hookID string) {
	endpointCacheKey := helper.EndpointCacheKey(hookID)
	if _, err := cache.AutoDelete[models.Endpoint](endpointCacheKey); err != nil {
		log.Warn("failed to invalidate endpoint cache", "hook_id", hookID, "error", err)
	}
}

func ListEndpoints(tenant_id string, page int, page_size int) *entities.Response {
	endpoints, err := db.GetAll[models.Endpoint](
		db.Equal("tenant_id", tenant_id),
		db.OrderBy("created_at", true),
		db.Page(page, page_size),
	)
	if err != nil {
		return exception.InternalServerError(fmt.Errorf("failed to list endpoints: %v", err)).ToResponse()
	}

	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(errors.New("failed to get plugin manager")).ToResponse()
	}

	// decrypt settings
	for i, endpoint := range endpoints {
		pluginInstallation, err := db.GetOne[models.PluginInstallation](
			db.Equal("plugin_id", endpoint.PluginID),
			db.Equal("tenant_id", tenant_id),
		)
		if err != nil {
			// use empty settings and declaration for uninstalled plugins
			endpoint.Settings = map[string]any{}
			endpoint.Declaration = &plugin_entities.EndpointProviderDeclaration{
				Settings:      []plugin_entities.ProviderConfig{},
				Endpoints:     []plugin_entities.EndpointDeclaration{},
				EndpointFiles: []string{},
			}
			endpoints[i] = endpoint
			continue
		}

		pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
			pluginInstallation.PluginUniqueIdentifier,
		)
		if err != nil {
			return exception.UniqueIdentifierError(
				fmt.Errorf("failed to parse plugin unique identifier: %v", err),
			).ToResponse()
		}

		pluginDeclaration, err := manager.GetDeclaration(
			pluginUniqueIdentifier,
			tenant_id,
			plugin_entities.PluginRuntimeType(pluginInstallation.RuntimeType),
		)
		if err != nil {
			return exception.InternalServerError(
				fmt.Errorf("failed to get plugin declaration: %v", err),
			).ToResponse()
		}

		if pluginDeclaration.Endpoint == nil {
			return exception.NotFoundError(errors.New("plugin does not have an endpoint")).ToResponse()
		}

		decryptedSettings, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
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
				Config:    pluginDeclaration.Endpoint.Settings,
			},
		})
		if err != nil {
			return exception.InternalServerError(
				fmt.Errorf("failed to decrypt settings: %v", err),
			).ToResponse()
		}

		// mask settings
		decryptedSettings = encryption.MaskConfigCredentials(decryptedSettings, pluginDeclaration.Endpoint.Settings)

		endpoint.Settings = decryptedSettings
		endpoint.Declaration = pluginDeclaration.Endpoint

		endpoints[i] = endpoint
	}

	return entities.NewSuccessResponse(endpoints)
}

func ListPluginEndpoints(tenant_id string, plugin_id string, page int, page_size int) *entities.Response {
	endpoints, err := db.GetAll[models.Endpoint](
		db.Equal("plugin_id", plugin_id),
		db.Equal("tenant_id", tenant_id),
		db.OrderBy("created_at", true),
		db.Page(page, page_size),
	)
	if err != nil {
		return exception.InternalServerError(
			fmt.Errorf("failed to list endpoints: %v", err),
		).ToResponse()
	}

	manager := plugin_manager.Manager()
	if manager == nil {
		return exception.InternalServerError(
			errors.New("failed to get plugin manager"),
		).ToResponse()
	}

	// decrypt settings
	for i, endpoint := range endpoints {
		// get installation
		pluginInstallation, err := db.GetOne[models.PluginInstallation](
			db.Equal("plugin_id", plugin_id),
			db.Equal("tenant_id", tenant_id),
		)
		if err != nil {
			return exception.NotFoundError(
				fmt.Errorf("failed to find plugin installation: %v", err),
			).ToResponse()
		}

		pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
			pluginInstallation.PluginUniqueIdentifier,
		)

		if err != nil {
			return exception.UniqueIdentifierError(
				fmt.Errorf("failed to parse plugin unique identifier: %v", err),
			).ToResponse()
		}

		pluginDeclaration, err := manager.GetDeclaration(
			pluginUniqueIdentifier,
			tenant_id,
			plugin_entities.PluginRuntimeType(pluginInstallation.RuntimeType),
		)
		if err != nil {
			return exception.InternalServerError(
				fmt.Errorf("failed to get plugin declaration: %v", err),
			).ToResponse()
		}

		if pluginDeclaration.Endpoint == nil {
			return exception.NotFoundError(errors.New("plugin does not have an endpoint")).ToResponse()
		}

		decryptedSettings, err := manager.BackwardsInvocation().InvokeEncrypt(&dify_invocation.InvokeEncryptRequest{
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
				Config:    pluginDeclaration.Endpoint.Settings,
			},
		})
		if err != nil {
			return exception.InternalServerError(
				fmt.Errorf("failed to decrypt settings: %v", err),
			).ToResponse()
		}

		// mask settings
		decryptedSettings = encryption.MaskConfigCredentials(decryptedSettings, pluginDeclaration.Endpoint.Settings)

		endpoint.Settings = decryptedSettings
		endpoint.Declaration = pluginDeclaration.Endpoint

		endpoints[i] = endpoint
	}

	return entities.NewSuccessResponse(endpoints)
}
