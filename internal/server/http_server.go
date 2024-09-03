package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func (app *App) server(config *app.Config) func() {
	engine := gin.Default()
	engine.GET("/health/check", controllers.HealthCheck)

	app.pluginInvokeGroup(engine.Group("/plugin"), config)
	app.remoteDebuggingGroup(engine.Group("/plugin/debugging"), config)
	app.endpointGroup(engine.Group("/e"), config)
	app.awsLambdaTransactionGroup(engine.Group("/backwards-invocation"), config)
	app.endpointManagementGroup(engine.Group("/endpoint"))

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

func (app *App) pluginInvokeGroup(group *gin.RouterGroup, config *app.Config) {
	group.Use(CheckingKey(config.PluginInnerApiKey))
	group.Use(app.RedirectPluginInvoke())
	group.Use(app.InitClusterID())

	group.POST("/tool/invoke", controllers.InvokeTool(config))
	group.POST("/tool/validate_credentials", controllers.ValidateToolCredentials(config))
	group.POST("/llm/invoke", controllers.InvokeLLM(config))
	group.POST("/text_embedding/invoke", controllers.InvokeTextEmbedding(config))
	group.POST("/rerank/invoke", controllers.InvokeRerank(config))
	group.POST("/tts/invoke", controllers.InvokeTTS(config))
	group.POST("/speech2text/invoke", controllers.InvokeSpeech2Text(config))
	group.POST("/moderation/invoke", controllers.InvokeModeration(config))
	group.POST("/model/validate_provider_credentials", controllers.ValidateProviderCredentials(config))
	group.POST("/model/validate_model_credentials", controllers.ValidateModelCredentials(config))
}

func (app *App) remoteDebuggingGroup(group *gin.RouterGroup, config *app.Config) {
	if config.PluginRemoteInstallingEnabled {
		group.POST("/key", CheckingKey(config.PluginInnerApiKey), controllers.GetRemoteDebuggingKey)
	}
}

func (app *App) endpointGroup(group *gin.RouterGroup, config *app.Config) {
	if config.PluginEndpointEnabled {
		group.HEAD("/:hook_id/*path", app.Endpoint())
		group.POST("/:hook_id/*path", app.Endpoint())
		group.GET("/:hook_id/*path", app.Endpoint())
		group.PUT("/:hook_id/*path", app.Endpoint())
		group.DELETE("/:hook_id/*path", app.Endpoint())
		group.OPTIONS("/:hook_id/*path", app.Endpoint())
	}
}

func (appRef *App) awsLambdaTransactionGroup(group *gin.RouterGroup, config *app.Config) {
	if config.Platform == app.PLATFORM_AWS_LAMBDA {
		appRef.aws_transaction_handler = transaction.NewAWSTransactionHandler(
			time.Duration(config.MaxAWSLambdaTransactionTimeout) * time.Second,
		)
		group.POST(
			"/transaction",
			service.HandleAWSPluginTransaction(appRef.aws_transaction_handler),
		)
	}
}

func (app *App) endpointManagementGroup(group *gin.RouterGroup) {
	group.POST("/setup", controllers.SetupEndpoint)
	group.POST("/remove", controllers.RemoveEndpoint)
	group.GET("/list", controllers.ListEndpoints)
}
