package server

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

func InvokeTool(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[plugin_entities.InvokeToolRequest]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeTool(&itr, c)
		},
	)
}
