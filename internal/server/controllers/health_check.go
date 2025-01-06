package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/manifest"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func HealthCheck(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"pool_status": routine.FetchRoutineStatus(),
			"version":     manifest.VersionX,
			"build_time":  manifest.BuildTimeX,
			"platform":    app.Platform,
		})
	}
}
