package server

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

// DifyPlugin supports register and use webhook to improve the plugin's functionality
// you can use it to do some magics, looking forward to your imagination, Ciallo～(∠·ω< )⌒
// - Yeuoly

// WebhookHandler is a function type that can be used to handle webhook requests
type WebhookHandler func(ctx *gin.Context, hook_id string, path string)

func (app *App) Webhook() func(c *gin.Context) {
	return func(c *gin.Context) {
		hook_id := c.Param("hook_id")
		path := c.Param("path")

		if app.webhook_handler != nil {
			app.webhook_handler(c, hook_id, path)
		} else {
			app.WebhookHandler(c, hook_id, path)
		}
	}
}

func (app *App) WebhookHandler(ctx *gin.Context, hook_id string, path string) {
	webhook, err := db.GetOne[models.Webhook](
		db.Equal("hook_id", hook_id),
	)

	if err == db.ErrDatabaseNotFound {
		ctx.JSON(404, gin.H{"error": "webhook not found"})
		return
	}

	if err != nil {
		log.Error("get webhook error %v", err)
		ctx.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	// check if plugin exists in current node
	if !app.cluster.IsPluginNoCurrentNode(webhook.PluginID) {
		app.Redirect(ctx, webhook.PluginID)
	} else {
		service.Webhook(ctx, &webhook, path)
	}
}
