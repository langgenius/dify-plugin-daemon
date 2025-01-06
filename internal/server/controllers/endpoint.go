package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func SetupEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(
		request struct {
			PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
			TenantID               string                                 `uri:"tenant_id" validate:"required"`
			UserID                 string                                 `json:"user_id" validate:"required"`
			Settings               map[string]any                         `json:"settings" validate:"omitempty"`
			Name                   string                                 `json:"name" validate:"required"`
		},
	) {
		tenantId := request.TenantID
		userId := request.UserID
		settings := request.Settings
		pluginUniqueIdentifier := request.PluginUniqueIdentifier
		name := request.Name

		ctx.JSON(200, service.SetupEndpoint(
			tenantId, userId, pluginUniqueIdentifier, name, settings,
		))
	})
}

func ListEndpoints(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		Page     int    `form:"page" validate:"required"`
		PageSize int    `form:"page_size" validate:"required,max=100"`
	}) {
		tenantId := request.TenantID
		page := request.Page
		pageSize := request.PageSize

		ctx.JSON(200, service.ListEndpoints(tenantId, page, pageSize))
	})
}

func ListPluginEndpoints(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		PluginID string `form:"plugin_id" validate:"required"`
		Page     int    `form:"page" validate:"required"`
		PageSize int    `form:"page_size" validate:"required,max=100"`
	}) {
		tenantId := request.TenantID
		pluginId := request.PluginID
		page := request.Page
		pageSize := request.PageSize

		ctx.JSON(200, service.ListPluginEndpoints(tenantId, pluginId, page, pageSize))
	})
}

func RemoveEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		EndpointID string `json:"endpoint_id" validate:"required"`
		TenantID   string `uri:"tenant_id" validate:"required"`
	}) {
		endpointId := request.EndpointID
		tenantId := request.TenantID

		ctx.JSON(200, service.RemoveEndpoint(endpointId, tenantId))
	})
}

func UpdateEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		EndpointID string         `json:"endpoint_id" validate:"required"`
		TenantID   string         `uri:"tenant_id" validate:"required"`
		UserID     string         `json:"user_id" validate:"required"`
		Settings   map[string]any `json:"settings" validate:"omitempty"`
		Name       string         `json:"name" validate:"required"`
	}) {
		tenantId := request.TenantID
		userId := request.UserID
		endpointId := request.EndpointID
		settings := request.Settings
		name := request.Name

		ctx.JSON(200, service.UpdateEndpoint(endpointId, tenantId, userId, name, settings))
	})
}

func EnableEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		EndpointID string `json:"endpoint_id" validate:"required"`
		TenantID   string `uri:"tenant_id" validate:"required"`
	}) {
		tenantId := request.TenantID
		endpointId := request.EndpointID

		ctx.JSON(200, service.EnableEndpoint(endpointId, tenantId))
	})
}

func DisableEndpoint(ctx *gin.Context) {
	BindRequest(ctx, func(request struct {
		EndpointID string `json:"endpoint_id" validate:"required"`
		TenantID   string `uri:"tenant_id" validate:"required"`
	}) {
		tenantId := request.TenantID
		endpointId := request.EndpointID

		ctx.JSON(200, service.DisableEndpoint(endpointId, tenantId))
	})
}
