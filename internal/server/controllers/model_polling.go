package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func StartPolling() gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestStartPolling]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(c, func(itr request) {
			service.StartPolling(&itr, c)
		})
	}
}

func CheckPolling() gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestCheckPolling]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(c, func(itr request) {
			service.CheckPolling(&itr, c)
		})
	}
}
