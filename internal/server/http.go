package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func (app *App) server(config *app.Config) {
	engine := gin.Default()

	engine.GET("/health/check", controllers.HealthCheck)

	engine.POST(
		"/plugin/tool/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeTool,
	)
	engine.POST(
		"/plugin/tool/validate_credentials",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.ValidateToolCredentials,
	)
	engine.POST(
		"/plugin/llm/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeLLM,
	)
	engine.POST(
		"/plugin/text_embedding/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeTextEmbedding,
	)
	engine.POST(
		"/plugin/rerank/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeRerank,
	)
	engine.POST(
		"/plugin/tts/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeTTS,
	)
	engine.POST(
		"/plugin/speech2text/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeSpeech2Text,
	)
	engine.POST(
		"/plugin/moderation/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.InvokeModeration,
	)
	engine.POST(
		"/plugin/model/validate_provider_credentials",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.ValidateProviderCredentials,
	)
	engine.POST(
		"/plugin/model/validate_model_credentials",
		CheckingKey(config.PluginInnerApiKey),
		app.Redirect(),
		controllers.ValidateModelCredentials,
	)

	if config.PluginRemoteInstallingEnabled {
		engine.POST(
			"/plugin/debugging/key",
			CheckingKey(config.PluginInnerApiKey),
			app.Redirect(),
			controllers.GetRemoteDebuggingKey,
		)
	}

	engine.Run(fmt.Sprintf(":%d", config.ServerPort))
}
