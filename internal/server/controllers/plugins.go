package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func GetAsset(c *gin.Context) {
	plugin_manager := plugin_manager.Manager()
	asset, err := plugin_manager.GetAsset(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", asset)
}

func InstallPluginFromPkg(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		dify_pkg_file_header, err := c.FormFile("dify_pkg")
		if err != nil {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-400, err.Error()))
			return
		}

		tenant_id := c.Param("tenant_id")
		if tenant_id == "" {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-400, "Tenant ID is required"))
			return
		}

		if dify_pkg_file_header.Size > app.MaxPluginPackageSize {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-413, "File size exceeds the maximum limit"))
			return
		}

		dify_pkg_file, err := dify_pkg_file_header.Open()
		if err != nil {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-500, err.Error()))
			return
		}
		defer dify_pkg_file.Close()

		service.InstallPluginFromPkg(c, tenant_id, dify_pkg_file)
	}
}

func InstallPluginFromIdentifier(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		BindRequest(c, func(request struct {
			TenantID               string                                 `uri:"tenant_id" binding:"required"`
			PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifier" binding:"required" validate:"plugin_unique_identifier"`
		}) {
			c.JSON(http.StatusOK, service.InstallPluginFromIdentifier(c, request.TenantID, request.PluginUniqueIdentifier))
		})
	}
}

func UninstallPlugin(c *gin.Context) {
}

func ListPlugins(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" binding:"required"`
		Page     int    `form:"page" binding:"required,min=1"`
		PageSize int    `form:"page_size" binding:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.ListPlugins(request.TenantID, request.Page, request.PageSize))
	})
}

func ListModels(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" binding:"required"`
		Page     int    `form:"page" binding:"required,min=1"`
		PageSize int    `form:"page_size" binding:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.ListModels(request.TenantID, request.Page, request.PageSize))
	})
}

func ListTools(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" binding:"required"`
		Page     int    `form:"page" binding:"required,min=1"`
		PageSize int    `form:"page_size" binding:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.ListTools(request.TenantID, request.Page, request.PageSize))
	})
}
