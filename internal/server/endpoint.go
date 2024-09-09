package server

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

// DifyPlugin supports register and use endpoint to improve the plugin's functionality
// you can use it to do some magics, looking forward to your imagination, Ciallo～(∠·ω< )⌒
// - Yeuoly

// EndpointHandler is a function type that can be used to handle endpoint requests
type EndpointHandler func(ctx *gin.Context, hook_id string, path string)

func (app *App) Endpoint() func(c *gin.Context) {
	return func(c *gin.Context) {
		hook_id := c.Param("hook_id")
		path := c.Param("path")

		if app.endpoint_handler != nil {
			app.endpoint_handler(c, hook_id, path)
		} else {
			app.EndpointHandler(c, hook_id, path)
		}
	}
}

func (app *App) EndpointHandler(ctx *gin.Context, hook_id string, path string) {
	endpoint, err := db.GetOne[models.Endpoint](
		db.Equal("hook_id", hook_id),
	)
	if err == db.ErrDatabaseNotFound {
		ctx.JSON(404, gin.H{"error": "endpoint not found"})
		return
	}

	if err != nil {
		log.Error("get endpoint error %v", err)
		ctx.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	// get plugin installation
	plugin_installation, err := db.GetOne[models.PluginInstallation](
		db.Equal("plugin_id", endpoint.PluginID),
		db.Equal("tenant_id", endpoint.TenantID),
	)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "plugin installation not found"})
		return
	}

	// check if plugin exists in current node
	if !app.cluster.IsPluginNoCurrentNode(
		plugin_entities.PluginUniqueIdentifier(plugin_installation.PluginUniqueIdentifier),
	) {
		app.redirectPluginInvokeByPluginIdentifier(ctx, plugin_entities.PluginUniqueIdentifier(
			plugin_installation.PluginUniqueIdentifier,
		))
	} else {
		service.Endpoint(ctx, &endpoint, &plugin_installation, path)
	}
}
