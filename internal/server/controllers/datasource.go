package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
)

func ListDatasources(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		Page     int    `form:"page" validate:"required,min=1"`
		PageSize int    `form:"page_size" validate:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.ListDatasources(request.TenantID, request.Page, request.PageSize))
	})
}

func GetDatasource(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		PluginID string `form:"plugin_id" validate:"required"`
		Provider string `form:"provider" validate:"required"`
	}) {
		c.JSON(http.StatusOK, service.GetDatasource(request.TenantID, request.PluginID, request.Provider))
	})
}
