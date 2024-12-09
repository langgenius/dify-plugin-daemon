package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
)

func InvokeAgent(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeAgent]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(
			c,
			func(itr request) {
				service.InvokeAgent(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}
