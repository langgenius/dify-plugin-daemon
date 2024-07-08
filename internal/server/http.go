package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func server(config *app.Config) {
	engine := gin.Default()

	engine.GET("/health/check", HealthCheck)
	engine.POST("/plugin/tool/invoke", CheckingKey(config.PluginInnerApiKey), InvokeTool)

	engine.Run(fmt.Sprintf(":%d", config.SERVER_PORT))
}
