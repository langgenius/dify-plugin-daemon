package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
)

func GetAsset(c *gin.Context) {
	plugin_manager := plugin_manager.GetGlobalPluginManager()
	asset, err := plugin_manager.GetAsset(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", asset)
}

func InstallPlugin(c *gin.Context) {
}

func UninstallPlugin(c *gin.Context) {
}

func ListPlugins(c *gin.Context) {
}
