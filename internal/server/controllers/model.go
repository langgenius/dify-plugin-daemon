package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
)

type listModelPluginBindingsRequest struct {
	TenantID string `uri:"tenant_id" validate:"required"`
}

func ListModelPluginBindings(c *gin.Context) {
	BindRequest(c, func(request listModelPluginBindingsRequest) {
		c.JSON(http.StatusOK, service.ListModelPluginBindings(request.TenantID))
	})
}

func ListModels(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		Page     int    `form:"page" validate:"required,min=1"`
		PageSize int    `form:"page_size" validate:"required,min=1,max=256"`
	}) {
		JSONResponse(c, service.ListModels(request.TenantID, request.Page, request.PageSize))
	})
}
