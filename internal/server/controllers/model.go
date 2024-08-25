package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
)

func InvokeLLM(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.InvokeLLM(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func InvokeTextEmbedding(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeTextEmbedding]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.InvokeTextEmbedding(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func InvokeRerank(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeRerank]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.InvokeRerank(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func InvokeTTS(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeTTS]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.InvokeTTS(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func InvokeSpeech2Text(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeSpeech2Text]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.InvokeSpeech2Text(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func InvokeModeration(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeModeration]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.InvokeModeration(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func ValidateProviderCredentials(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateProviderCredentials]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.ValidateProviderCredentials(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}

func ValidateModelCredentials(config *app.Config) gin.HandlerFunc {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateModelCredentials]

	return func(c *gin.Context) {
		BindRequest[request](
			c,
			func(itr request) {
				service.ValidateModelCredentials(&itr, c, config.PluginMaxExecutionTimeout)
			},
		)
	}
}
