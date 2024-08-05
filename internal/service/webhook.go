package service

import (
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
)

func Webhook(ctx *gin.Context, webhook *models.Webhook, path string) {
	req := ctx.Request

	var buffer bytes.Buffer
	err := req.Write(&buffer)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}

	// fetch plugin
	manager := plugin_manager.GetGlobalPluginManager()
	runtime := manager.Get(webhook.PluginID)
	if runtime == nil {
		ctx.JSON(404, gin.H{"error": "plugin not found"})
		return
	}

	session := session_manager.NewSession(webhook.TenantID, "", webhook.PluginID)
	defer session.Close()

	session.BindRuntime(runtime)

	// TODO: handle webhook
}
