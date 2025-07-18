package server

import (
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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

		tenantId := ctx.Param("tenant_id")
		if tenantId == "" {
			ctx.AbortWithStatusJSON(400, exception.BadRequestError(errors.New("tenant_id is required")).ToResponse())
			return
		}

		// fetch plugin installation with caching
		cacheKey := helper.PluginInstallationCacheKey(pluginId, tenantId)
		installation, err := cache.AutoGetWithGetter(
			cacheKey,
			func() (*models.PluginInstallation, error) {
				inst, err := db.GetOne[models.PluginInstallation](
					db.Equal("tenant_id", tenantId),
					db.Equal("plugin_id", pluginId),
				)
				if err != nil {
					return nil, err
				}
				return &inst, nil
			},
		)

		if err == db.ErrDatabaseNotFound {
			ctx.AbortWithStatusJSON(404, exception.ErrPluginNotFound().ToResponse())
			return
		}

		if err != nil {
			ctx.AbortWithStatusJSON(500, exception.InternalServerError(err).ToResponse())
			return
		}

		identity, err := plugin_entities.NewPluginUniqueIdentifier(installation.PluginUniqueIdentifier)
		if err != nil {
			ctx.AbortWithStatusJSON(400, exception.UniqueIdentifierError(err).ToResponse())
			return
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

		// check if plugin in current node
		if ok, originalError := app.cluster.IsPluginOnCurrentNode(identity); !ok {
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
		ctx.AbortWithStatusJSON(
			500,
			exception.InternalServerError(
				errors.New("failed to fetch plugin available nodes, "+originalError.Error()+", "+err.Error()),
			).ToResponse(),
		)
		return
	} else if len(nodes) == 0 {
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
		log.Error("redirect request failed: %s", err.Error())
		ctx.AbortWithStatusJSON(
			500,
			exception.InternalServerError(errors.New("redirect request failed: "+err.Error())).ToResponse(),
		)
		return
	}

	// set status code
	ctx.Writer.WriteHeader(statusCode)

	// set header
	for key, values := range header {
		for _, value := range values {
			ctx.Writer.Header().Set(key, value)
		}
	}

	for {
		buf := make([]byte, 1024)
		n, err := body.Read(buf)
		if err != nil && err != io.EOF {
			break
		} else if err != nil {
			ctx.Writer.Write(buf[:n])
			ctx.Writer.Flush()
			break
		}

		if n > 0 {
			ctx.Writer.Write(buf[:n])
			ctx.Writer.Flush()
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
			ctx.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
			return
		}

		ctx.Next()
	}
}
