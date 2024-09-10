package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
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

func InstallPlugin(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		dify_pkg_file_header, err := c.FormFile("dify_pkg")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if dify_pkg_file_header.Size > app.MaxPluginPackageSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size exceeds the maximum limit"})
			return
		}

		dify_pkg_file, err := dify_pkg_file_header.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer dify_pkg_file.Close()

		service.InstallPlugin(c, dify_pkg_file)
	}
}

func UninstallPlugin(c *gin.Context) {
}

func ListPlugins(c *gin.Context) {
}
