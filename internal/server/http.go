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
	engine.POST("/plugin/tool/validate_credentials", CheckingKey(config.PluginInnerApiKey), ValidateToolCredentials)
	engine.POST("/plugin/llm/invoke", CheckingKey(config.PluginInnerApiKey), InvokeLLM)
	engine.POST("/plugin/text_embedding/invoke", CheckingKey(config.PluginInnerApiKey), InvokeTextEmbedding)
	engine.POST("/plugin/rerank/invoke", CheckingKey(config.PluginInnerApiKey), InvokeRerank)
	engine.POST("/plugin/tts/invoke", CheckingKey(config.PluginInnerApiKey), InvokeTTS)
	engine.POST("/plugin/speech2text/invoke", CheckingKey(config.PluginInnerApiKey), InvokeSpeech2Text)
	engine.POST("/plugin/moderation/invoke", CheckingKey(config.PluginInnerApiKey), InvokeModeration)
	engine.POST("/plugin/model/validate_provider_credentials", CheckingKey(config.PluginInnerApiKey), ValidateProviderCredentials)
	engine.POST("/plugin/model/validate_model_credentials", CheckingKey(config.PluginInnerApiKey), ValidateModelCredentials)

	engine.Run(fmt.Sprintf(":%d", config.SERVER_PORT))
}
