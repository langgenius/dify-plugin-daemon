package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
)

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
