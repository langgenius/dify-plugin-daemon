package server

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/metrics"
)

func CheckingKey(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get header X-Api-Key
		if c.GetHeader(constants.X_API_KEY) != key {
			c.AbortWithStatusJSON(401, exception.UnauthorizedError().ToResponse())
			return
		}

		c.Next()
	}
}

func (app *App) FetchPluginInstallation() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pluginId := ctx.Request.Header.Get(constants.X_PLUGIN_ID)
		if pluginId == "" {
			ctx.AbortWithStatusJSON(400, exception.BadRequestError(errors.New("plugin_id is required")).ToResponse())
			return
		}

		identifier := ctx.GetHeader(constants.PluginUniqueIdentifier)
		var identity plugin_entities.PluginUniqueIdentifier
		if identifier != "" {
			var err error
			identity, err = plugin_entities.NewPluginUniqueIdentifier(identifier)
			if err != nil {
				ctx.AbortWithStatusJSON(400, exception.UniqueIdentifierError(err).ToResponse())
				return
			}
			if identity.PluginID() != pluginId {
				ctx.AbortWithStatusJSON(400, exception.BadRequestError(errors.New("plugin unique identifier does not match X-Plugin-ID")).ToResponse())
				return
			}
		}

		tenantId := ctx.Param("tenant_id")
		if tenantId == "" {
			ctx.AbortWithStatusJSON(400, exception.BadRequestError(errors.New("tenant_id is required")).ToResponse())
			return
		}

		installationQuery := []db.GenericQuery{
			db.Equal("tenant_id", tenantId),
			db.Equal("plugin_id", pluginId),
		}
		fetchInstallation := func() (*models.PluginInstallation, error) {
			inst, err := db.GetOne[models.PluginInstallation](installationQuery...)
			if err != nil {
				return nil, err
			}
			return &inst, nil
		}

		var installation *models.PluginInstallation
		var err error
		if identifier != "" {
			// Authorize the plugin independently of its installed version so an
			// upgrade does not reject a prepared request. Bypass stale uninstall caches.
			installation, err = fetchInstallation()
		} else {
			cacheKey := helper.PluginInstallationCacheKey(pluginId, tenantId)
			installation, err = cache.AutoGetWithGetter(cacheKey, fetchInstallation)
		}

		if errors.Is(err, db.ErrDatabaseNotFound) {
			ctx.AbortWithStatusJSON(404, exception.ErrPluginNotFound().ToResponse())
			return
		}

		if err != nil {
			ctx.AbortWithStatusJSON(500, exception.InternalServerError(err).ToResponse())
			return
		}

		// Only legacy clients resolve the installed version through the plugin ID.
		if identifier == "" {
			identity, err = plugin_entities.NewPluginUniqueIdentifier(installation.PluginUniqueIdentifier)
			if err != nil {
				ctx.AbortWithStatusJSON(400, exception.UniqueIdentifierError(err).ToResponse())
				return
			}
		}
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_INSTALLATION, *installation)
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, identity)
		ctx.Next()
	}
}

// RedirectPluginInvoke redirects the request to the correct cluster node
func (app *App) RedirectPluginInvoke() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get plugin unique identifier
		identityAny, ok := ctx.Get(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER)
		if !ok {
			ctx.AbortWithStatusJSON(
				500,
				exception.InternalServerError(errors.New("plugin unique identifier not found")).ToResponse(),
			)
			return
		}

		identity, ok := identityAny.(plugin_entities.PluginUniqueIdentifier)
		if !ok {
			ctx.AbortWithStatusJSON(
				500,
				exception.InternalServerError(errors.New("failed to parse plugin unique identifier")).ToResponse(),
			)
			return
		}

		runtimeType := plugin_entities.PluginRuntimeType("")
		if installationAny, ok := ctx.Get(constants.CONTEXT_KEY_PLUGIN_INSTALLATION); ok {
			installation, ok := installationAny.(models.PluginInstallation)
			if !ok {
				ctx.AbortWithStatusJSON(
					500,
					exception.InternalServerError(errors.New("failed to parse plugin installation")).ToResponse(),
				)
				return
			}
			runtimeType = plugin_entities.PluginRuntimeType(installation.RuntimeType)
		}

		// check if plugin in current node
		if needRedirecting, originalError := app.pluginManager.NeedRedirecting(
			identity,
			runtimeType,
		); needRedirecting {
			app.redirectPluginInvokeByPluginIdentifier(ctx, identity, originalError)
			ctx.Abort()
		} else {
			ctx.Next()
		}
	}
}

func (app *App) redirectPluginInvokeByPluginIdentifier(
	ctx *gin.Context,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	originalError error,
) {
	// try find the correct node
	nodes, err := app.cluster.FetchPluginAvailableNodesById(plugin_unique_identifier.String())
	if err != nil {
		log.Error("Failed to fetch plugin nodes by id", "error", err)
		ctx.AbortWithStatusJSON(
			500,
			exception.InternalServerError(
				errors.New("failed to fetch plugin available nodes, "+originalError.Error()+", "+err.Error()),
			).ToResponse(),
		)
		return
	} else if len(nodes) == 0 {
		log.Error("no plugin available nodes found", "plugin", plugin_unique_identifier.String())
		ctx.AbortWithStatusJSON(
			404,
			exception.InternalServerError(
				errors.New("no available node, "+originalError.Error()),
			).ToResponse(),
		)
		return
	}

	// redirect to the correct node
	nodeId := nodes[0]
	statusCode, header, body, err := app.cluster.RedirectRequest(nodeId, ctx.Request)
	if err != nil {
		log.Error("redirect request failed", "error", err)
		ctx.AbortWithStatusJSON(
			500,
			exception.InternalServerError(errors.New("redirect request failed: "+err.Error())).ToResponse(),
		)
		return
	}

	writeRedirectedResponse(ctx, statusCode, header, body)
}

func writeRedirectedResponse(ctx *gin.Context, statusCode int, header http.Header, body io.ReadCloser) {
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Error("body close failed", "error", err)
		}
	}(body)

	setHeaders := func() {
		if !ctx.Writer.Written() {
			ctx.Status(statusCode)
			for key, values := range header {
				ctx.Writer.Header()[key] = values
			}
		}
	}
	buffer := make([]byte, 32*1024)
	for {
		n, err := body.Read(buffer)
		if n > 0 {
			setHeaders()
			if _, writeErr := ctx.Writer.Write(buffer[:n]); writeErr != nil {
				log.Error("failed to write redirected response body", "error", writeErr)
				panic(http.ErrAbortHandler)
			}
			ctx.Writer.Flush()
		}
		if err == io.EOF {
			setHeaders()
			return
		}
		if err != nil {
			failure := exception.InternalServerError(err).ToResponse()
			if ctx.Writer.Written() {
				panic(http.ErrAbortHandler)
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, failure)
			return
		}
	}
}

func (app *App) InitClusterID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(constants.CONTEXT_KEY_CLUSTER_ID, app.cluster.ID())
		ctx.Next()
	}
}

func (app *App) AdminAPIKey(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(constants.X_ADMIN_API_KEY) != key {
			ctx.AbortWithStatusJSON(
				401,
				exception.UnauthorizedError().ToResponse())
			return
		}
		ctx.Next()
	}
}

func (app *App) FetchPluginDirect() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		identifier := ctx.Request.Header.Get(constants.PluginUniqueIdentifier)
		if identifier == "" {
			ctx.AbortWithStatusJSON(
				400,
				exception.BadRequestError(errors.New("X-Plugin-Unique-Identifier header is required")).ToResponse())
			return
		}
		pluginID, err := plugin_entities.NewPluginUniqueIdentifier(identifier)
		if err != nil {
			ctx.AbortWithStatusJSON(400,
				exception.UniqueIdentifierError(err).ToResponse(),
			)
			return
		}
		ctx.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, pluginID)
		ctx.Next()
	}
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
