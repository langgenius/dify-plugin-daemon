package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func (app *App) server(config *app.Config) func() {
	engine := gin.Default()

	engine.GET("/health/check", controllers.HealthCheck)

	engine.POST(
		"/plugin/tool/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeTool,
	)
	engine.POST(
		"/plugin/tool/validate_credentials",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.ValidateToolCredentials,
	)
	engine.POST(
		"/plugin/llm/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeLLM,
	)
	engine.POST(
		"/plugin/text_embedding/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeTextEmbedding,
	)
	engine.POST(
		"/plugin/rerank/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeRerank,
	)
	engine.POST(
		"/plugin/tts/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeTTS,
	)
	engine.POST(
		"/plugin/speech2text/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeSpeech2Text,
	)
	engine.POST(
		"/plugin/moderation/invoke",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.InvokeModeration,
	)
	engine.POST(
		"/plugin/model/validate_provider_credentials",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.ValidateProviderCredentials,
	)
	engine.POST(
		"/plugin/model/validate_model_credentials",
		CheckingKey(config.PluginInnerApiKey),
		app.RedirectPluginInvoke(),
		controllers.ValidateModelCredentials,
	)

	if config.PluginRemoteInstallingEnabled {
		engine.POST(
			"/plugin/debugging/key",
			CheckingKey(config.PluginInnerApiKey),
			app.RedirectPluginInvoke(),
			controllers.GetRemoteDebuggingKey,
		)
	}

	if config.PluginWebhookEnabled {
		engine.HEAD("/webhook/:hook_id/*path", app.Webhook())
		engine.POST("/webhook/:hook_id/*path", app.Webhook())
		engine.GET("/webhook/:hook_id/*path", app.Webhook())
		engine.PUT("/webhook/:hook_id/*path", app.Webhook())
		engine.DELETE("/webhook/:hook_id/*path", app.Webhook())
		engine.OPTIONS("/webhook/:hook_id/*path", app.Webhook())
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerPort),
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panic("listen: %s\n", err)
		}
	}()

	return func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Panic("Server Shutdown: %s\n", err)
		}
	}
}
