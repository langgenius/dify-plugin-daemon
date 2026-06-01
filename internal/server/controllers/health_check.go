package controllers

import (
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/pip"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/manifest"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

var (
	activeRequests         int32 = 0 // how many requests are active
	activeDispatchRequests int32 = 0 // how many plugin dispatching requests are active
)

func CollectActiveRequests() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		atomic.AddInt32(&activeRequests, 1)
		ctx.Next()
		atomic.AddInt32(&activeRequests, -1)
	}
}

func CollectActiveDispatchRequests() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		atomic.AddInt32(&activeDispatchRequests, 1)
		ctx.Next()
		atomic.AddInt32(&activeDispatchRequests, -1)
	}
}

func HealthCheck(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := gin.H{
			"status":                   "ok",
			"pool_status":              routine.FetchRoutineStatus(),
			"version":                  manifest.VersionX,
			"build_time":               manifest.BuildTimeX,
			"platform":                 app.Platform,
			"active_requests":          activeRequests,
			"active_dispatch_requests": activeDispatchRequests,
		}

		// Surface the selected PyPI mirror's connectivity as an informational
		// field. This never affects the response status code; it only serves as a
		// reminder when the effective mirror is unreachable.
		if result, ok := pip.GetResult(); ok {
			if status, found := result.SelectedStatus(); found {
				body["pypi"] = status
			} else {
				body["pypi"] = gin.H{"selected": result.Selected}
			}
		}

		c.JSON(200, body)
	}
}
