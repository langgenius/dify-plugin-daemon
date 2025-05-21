package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func DatasourceValidateCredentials(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateDatasourceCredentials]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(
			c,
			func(ipr request) {
				service.DatasourceValidateCredentials(
					&ipr,
					c,
					time.Duration(config.PluginMaxExecutionTimeout)*time.Second,
				)
			},
		)
	}
}

func DatasourceInvokeFirstStep(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeDatasourceFirstStep]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(c, func(ipr request) {
			service.DatasourceInvokeFirstStep(
				&ipr,
				c,
				time.Duration(config.PluginMaxExecutionTimeout)*time.Second,
			)
		})
	}
}

func DatasourceInvokeSecondStep(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeDatasourceSecondStep]

	return func(c *gin.Context) {
		BindPluginDispatchRequest(c, func(ipr request) {
			service.DatasourceInvokeSecondStep(
				&ipr,
				c,
				time.Duration(config.PluginMaxExecutionTimeout)*time.Second,
			)
		})
	}
}
