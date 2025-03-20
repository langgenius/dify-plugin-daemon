package server

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// DifyPlugin supports register and use endpoint to improve the plugin's functionality
// you can use it to do some magics, looking forward to your imagination, Ciallo～(∠·ω< )⌒
// - Yeuoly

// EndpointHandler is a function type that can be used to handle endpoint requests
type EndpointHandler func(ctx *gin.Context, hookId string, maxExecutionTime time.Duration, path string)

func (app *App) Endpoint(config *app.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		hookId := c.Param("hook_id")
		path := c.Param("path")

		if app.endpointHandler != nil {
			app.endpointHandler(c, hookId, time.Duration(config.PluginMaxExecutionTimeout)*time.Second, path)
		} else {
			app.EndpointHandler(c, hookId, time.Duration(config.PluginMaxExecutionTimeout)*time.Second, path)
		}
	}
}

func (app *App) EndpointHandler(ctx *gin.Context, hookId string, maxExecutionTime time.Duration, path string) {
	endpoint, err := db.GetCache[models.Endpoint](&db.GetCachePayload[models.Endpoint]{
		Getter: func() (*models.Endpoint, error) {
			v, err := db.GetOne[models.Endpoint](
				db.Equal("hook_id", hookId),
			)
			return &v, err
		},
		CacheKey: []db.KeyValuePair{
			{Key: "hook_id", Val: hookId},
		},
	})
	if err == db.ErrDatabaseNotFound {
		ctx.JSON(404, exception.BadRequestError(errors.New("endpoint not found")).ToResponse())
		return
	}

	if err != nil {
		log.Error("get endpoint error %v", err)
		ctx.JSON(500, exception.InternalServerError(errors.New("internal server error")).ToResponse())
		return
	}

	// get plugin installation
	pluginInstallation, err := db.GetCache[models.PluginInstallation](
		&db.GetCachePayload[models.PluginInstallation]{
			Getter: func() (*models.PluginInstallation, error) {
				v, err := db.GetOne[models.PluginInstallation](
					db.Equal("plugin_id", endpoint.PluginID),
					db.Equal("tenant_id", endpoint.TenantID),
				)
				return &v, err
			},
			CacheKey: []db.KeyValuePair{
				{Key: "plugin_id", Val: endpoint.PluginID},
				{Key: "tenant_id", Val: endpoint.TenantID},
			},
		})
	if err != nil {
		ctx.JSON(404, exception.BadRequestError(errors.New("plugin installation not found")).ToResponse())
		return
	}

	pluginUniqueIdentifier, err := plugin_entities.NewPluginUniqueIdentifier(
		pluginInstallation.PluginUniqueIdentifier,
	)
	if err != nil {
		ctx.JSON(400, exception.UniqueIdentifierError(
			errors.New("invalid plugin unique identifier"),
		).ToResponse())
		return
	}

	// check if plugin exists in current node
	if ok, originalError := app.cluster.IsPluginOnCurrentNode(pluginUniqueIdentifier); !ok {
		app.redirectPluginInvokeByPluginIdentifier(ctx, pluginUniqueIdentifier, originalError)
	} else {
		service.Endpoint(ctx, endpoint, pluginInstallation, maxExecutionTime, path)
	}
}
