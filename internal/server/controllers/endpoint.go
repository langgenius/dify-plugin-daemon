package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func SetupEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		PluginIdentity string         `json:"plugin_identity" binding:"required"`
		TenantID       string         `json:"tenant_id" binding:"required"`
		UserID         string         `json:"user_id" binding:"required"`
		Settings       map[string]any `json:"settings" binding:"omitempty"`
	}) {
		plugin_identity := request.PluginIdentity
		tenant_id := request.TenantID
		user_id := request.UserID
		settings := request.Settings

		ctx.JSON(200, service.SetupEndpoint(
			tenant_id, user_id, plugin_entities.PluginIdentity(plugin_identity), settings,
		))
	})
}

func ListEndpoints(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		TenantID string `form:"tenant_id" binding:"required"`
		Page     int    `form:"page" binding:"required"`
		PageSize int    `form:"page_size" binding:"required,max=100"`
	}) {
		tenant_id := request.TenantID
		page := request.Page
		page_size := request.PageSize

		ctx.JSON(200, service.ListEndpoints(tenant_id, page, page_size))
	})
}

func RemoveEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		PluginIdentity string `json:"plugin_identity"`
		TenantID       string `json:"tenant_id"`
	}) {
		plugin_identity := request.PluginIdentity
		tenant_id := request.TenantID

		ctx.JSON(200, service.RemoveEndpoint(plugin_identity, tenant_id))
	})
}
