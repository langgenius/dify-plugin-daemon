package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func server(config *app.Config) {
	engine := gin.Default()

	engine.GET("/health/check", controllers.HealthCheck)

	engine.POST("/plugin/tool/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeTool)
	engine.POST("/plugin/tool/validate_credentials", CheckingKey(config.PluginInnerApiKey), controllers.ValidateToolCredentials)
	engine.POST("/plugin/llm/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeLLM)
	engine.POST("/plugin/text_embedding/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeTextEmbedding)
	engine.POST("/plugin/rerank/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeRerank)
	engine.POST("/plugin/tts/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeTTS)
	engine.POST("/plugin/speech2text/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeSpeech2Text)
	engine.POST("/plugin/moderation/invoke", CheckingKey(config.PluginInnerApiKey), controllers.InvokeModeration)
	engine.POST("/plugin/model/validate_provider_credentials", CheckingKey(config.PluginInnerApiKey), controllers.ValidateProviderCredentials)
	engine.POST("/plugin/model/validate_model_credentials", CheckingKey(config.PluginInnerApiKey), controllers.ValidateModelCredentials)

	if config.PluginRemoteInstallingEnabled {
		engine.POST("/plugin/debugging/key", CheckingKey(config.PluginInnerApiKey), controllers.GetRemoteDebuggingKey)
	}

	engine.Run(fmt.Sprintf(":%d", config.ServerPort))
}
