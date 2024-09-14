package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func SetupEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(
		request struct {
			PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifier" binding:"required" validate:"plugin_unique_identifier"`
			TenantID               string                                 `json:"tenant_id" binding:"required"`
			UserID                 string                                 `json:"user_id" binding:"required"`
			Settings               map[string]any                         `json:"settings" binding:"omitempty"`
		},
	) {
		tenant_id := request.TenantID
		user_id := request.UserID
		settings := request.Settings
		plugin_unique_identifier := request.PluginUniqueIdentifier

		ctx.JSON(200, service.SetupEndpoint(
			tenant_id, user_id, plugin_unique_identifier, settings,
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
		EndpointID string `json:"endpoint_id" binding:"required"`
		TenantID   string `json:"tenant_id" binding:"required"`
	}) {
		endpoint_id := request.EndpointID
		tenant_id := request.TenantID

		ctx.JSON(200, service.RemoveEndpoint(endpoint_id, tenant_id))
	})
}

func EnableEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		EndpointID string `json:"endpoint_id" binding:"required"`
		TenantID   string `json:"tenant_id" binding:"required"`
	}) {
		tenant_id := request.TenantID
		endpoint_id := request.EndpointID

		ctx.JSON(200, service.EnableEndpoint(endpoint_id, tenant_id))
	})
}

func DisableEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		EndpointID string `json:"endpoint_id" binding:"required"`
		TenantID   string `json:"tenant_id" binding:"required"`
	}) {
		tenant_id := request.TenantID
		endpoint_id := request.EndpointID

		ctx.JSON(200, service.DisableEndpoint(endpoint_id, tenant_id))
	})
}
