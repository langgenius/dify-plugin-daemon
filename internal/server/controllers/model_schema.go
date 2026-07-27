package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func GetAIModelSchema(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestGetAIModelSchema]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(
			c,
			func(itr request) {
				service.GetAIModelSchema(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}
