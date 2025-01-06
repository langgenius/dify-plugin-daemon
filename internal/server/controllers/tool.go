package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func InvokeTool(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeTool]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(
			c,
			func(itr request) {
				service.InvokeTool(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func ValidateToolCredentials(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateToolCredentials]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(
			c,
			func(itr request) {
				service.ValidateToolCredentials(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func GetToolRuntimeParameters(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestGetToolRuntimeParameters]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(
			c,
			func(itr request) {
				service.GetToolRuntimeParameters(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func ListTools(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		Page     int    `form:"page" validate:"required,min=1"`
		PageSize int    `form:"page_size" validate:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.ListTools(request.TenantID, request.Page, request.PageSize))
	})
}

func GetTool(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		PluginID string `form:"plugin_id" validate:"required"`
		Provider string `form:"provider" validate:"required"`
	}) {
		c.JSON(http.StatusOK, service.GetTool(request.TenantID, request.PluginID, request.Provider))
	})
}

func CheckToolExistence(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID    string                              `uri:"tenant_id" validate:"required"`
		ProviderIDS []service.RequestCheckToolExistence `json:"provider_ids" validate:"required,dive"`
	}) {
		c.JSON(http.StatusOK, service.CheckToolExistence(request.TenantID, request.ProviderIDS))
	})
}
