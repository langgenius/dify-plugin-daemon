package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
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

	app.endpointGroup(engine.Group("/e"), config)
	app.awsLambdaTransactionGroup(engine.Group("/backwards-invocation"), config)
	app.pluginGroup(engine.Group("/plugin/:tenant_id"), config)

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

func (app *App) pluginGroup(group *gin.RouterGroup, config *app.Config) {
	group.Use(CheckingKey(config.ServerKey))

	app.remoteDebuggingGroup(group.Group("/debugging"), config)
	app.pluginDispatchGroup(group.Group("/dispatch"), config)
	app.pluginManagementGroup(group.Group("/management"), config)
	app.endpointManagementGroup(group.Group("/endpoint"))
	app.pluginAssetGroup(group.Group("/asset"))
}

func (app *App) pluginDispatchGroup(group *gin.RouterGroup, config *app.Config) {
	group.Use(app.FetchPluginInstallation())
	group.Use(app.RedirectPluginInvoke())
	group.Use(app.InitClusterID())

	group.POST("/tool/invoke", controllers.InvokeTool(config))
	group.POST("/tool/validate_credentials", controllers.ValidateToolCredentials(config))
	group.POST("/tool/get_runtime_parameters", controllers.GetToolRuntimeParameters(config))
	group.POST("/llm/invoke", controllers.InvokeLLM(config))
	group.POST("/llm/num_tokens", controllers.GetLLMNumTokens(config))
	group.POST("/text_embedding/invoke", controllers.InvokeTextEmbedding(config))
	group.POST("/text_embedding/num_tokens", controllers.GetTextEmbeddingNumTokens(config))
	group.POST("/rerank/invoke", controllers.InvokeRerank(config))
	group.POST("/tts/invoke", controllers.InvokeTTS(config))
	group.POST("/tts/model/voices", controllers.GetTTSModelVoices(config))
	group.POST("/speech2text/invoke", controllers.InvokeSpeech2Text(config))
	group.POST("/moderation/invoke", controllers.InvokeModeration(config))
	group.POST("/model/validate_provider_credentials", controllers.ValidateProviderCredentials(config))
	group.POST("/model/validate_model_credentials", controllers.ValidateModelCredentials(config))
	group.POST("/model/schema", controllers.GetAIModelSchema(config))
}

func (app *App) remoteDebuggingGroup(group *gin.RouterGroup, config *app.Config) {
	if config.PluginRemoteInstallingEnabled {
		group.POST("/key", CheckingKey(config.ServerKey), controllers.GetRemoteDebuggingKey)
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
	group.POST("/update", controllers.UpdateEndpoint)
	group.GET("/list", controllers.ListEndpoints)
	group.GET("/list/plugin", controllers.ListPluginEndpoints)
	group.POST("/enable", controllers.EnableEndpoint)
	group.POST("/disable", controllers.DisableEndpoint)
}

func (app *App) pluginManagementGroup(group *gin.RouterGroup, config *app.Config) {
	group.POST("/install/pkg", controllers.InstallPluginFromPkg(config))
	group.POST("/install/identifier", controllers.InstallPluginFromIdentifier(config))
	group.GET("/fetch/identifier", controllers.FetchPluginFromIdentifier)
	group.POST("/uninstall", controllers.UninstallPlugin)
	group.GET("/list", gzip.Gzip(gzip.DefaultCompression), controllers.ListPlugins)
	group.GET("/models", gzip.Gzip(gzip.DefaultCompression), controllers.ListModels)
	group.GET("/tools", gzip.Gzip(gzip.DefaultCompression), controllers.ListTools)
	group.GET("/tool", gzip.Gzip(gzip.DefaultCompression), controllers.GetTool)
}

func (app *App) pluginAssetGroup(group *gin.RouterGroup) {
	group.GET("/:id", gzip.Gzip(gzip.DefaultCompression), controllers.GetAsset)
}
