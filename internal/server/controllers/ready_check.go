package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
)

func ReadyCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		report := plugin_manager.Manager().Readiness()
		if report.Ready {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"ready":  true,
				"reason": report.Reason,
			})
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unready",
			"ready":  false,
			"reason": report.Reason,
			"detail": report.Plugins,
		})
	}
}
