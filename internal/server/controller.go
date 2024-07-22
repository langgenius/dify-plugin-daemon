package server

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
)

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

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

func InvokeLLM(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeLLM(&itr, c)
		},
	)
}

func InvokeTextEmbedding(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeTextEmbedding]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeTextEmbedding(&itr, c)
		},
	)
}

func InvokeRerank(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeRerank]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeRerank(&itr, c)
		},
	)
}

func InvokeTTS(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeTTS]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeTTS(&itr, c)
		},
	)
}

func InvokeSpeech2Text(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeSpeech2Text]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeSpeech2Text(&itr, c)
		},
	)
}

func InvokeModeration(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestInvokeModeration]

	BindRequest[request](
		c,
		func(itr request) {
			service.InvokeModeration(&itr, c)
		},
	)
}

func ValidateProviderCredentials(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateProviderCredentials]

	BindRequest[request](
		c,
		func(itr request) {
			service.ValidateProviderCredentials(&itr, c)
		},
	)
}

func ValidateModelCredentials(c *gin.Context) {
	type request = plugin_entities.InvokePluginRequest[requests.RequestValidateModelCredentials]

	BindRequest[request](
		c,
		func(itr request) {
			service.ValidateModelCredentials(&itr, c)
		},
	)
}
