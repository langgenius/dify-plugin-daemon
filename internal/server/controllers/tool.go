package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
)

func InvokeTool(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeTool]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeTool(&itr, c)
		},
	)
}

func ValidateToolCredentials(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateToolCredentials]

	BindRequest[request](
		c,
		func(itr request) {
			service.ValidateToolCredentials(&itr, c)
		},
	)
}
