package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// DifyPlugin supports register and use webhook to improve the plugin's functionality
// you can use it to do some magic things, look forward to your imagination Ciallo～(∠·ω< )⌒
// - Yeuoly

// WebhookHandler is a function type that can be used to handle webhook requests
type WebhookHandler func(hook_id string, path string)

func (app *App) Webhook() func(c *gin.Context) {
	return func(c *gin.Context) {
		hook_id := c.Param("hook_id")
		path := c.Param("path")

		if app.webhook_handler != nil {
			app.webhook_handler(hook_id, path)
		} else {
			app.WebhookHandler(hook_id, path)
		}
	}
}

func (app *App) WebhookHandler(hook_id string, path string) {
	fmt.Println(hook_id)
	fmt.Println(path)
}
