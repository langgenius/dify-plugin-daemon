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

	sentrygin "github.com/getsentry/sentry-go/gin"
)

// server starts a http server and returns a function to stop it
func (app *App) server(config *app.Config) func() {
	engine := gin.New()
	if *config.HealthApiLogEnabled {
		engine.Use(gin.Logger())
	} else {
		engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/health/check"},
		}))
	}
	engine.Use(gin.Recovery())
	engine.GET("/health/check", controllers.HealthCheck(config))

	endpointGroup := engine.Group("/e")
	awsLambdaTransactionGroup := engine.Group("/backwards-invocation")
	pluginGroup := engine.Group("/plugin/:tenant_id")
	pprofGroup := engine.Group("/debug/pprof")

	if config.SentryEnabled {
		// setup sentry for all groups
		sentryGroup := []*gin.RouterGroup{
			endpointGroup,
			awsLambdaTransactionGroup,
			pluginGroup,
		}
		for _, group := range sentryGroup {
			group.Use(sentrygin.New(sentrygin.Options{
				Repanic: true,
			}))
		}
	}

	app.endpointGroup(endpointGroup, config)
	app.awsLambdaTransactionGroup(awsLambdaTransactionGroup, config)
	app.pluginGroup(pluginGroup, config)
	app.pprofGroup(pprofGroup, config)

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
	group.POST("/agent_strategy/invoke", controllers.InvokeAgentStrategy(config))
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
	if config.PluginRemoteInstallingEnabled != nil && *config.PluginRemoteInstallingEnabled {
		group.POST("/key", CheckingKey(config.ServerKey), controllers.GetRemoteDebuggingKey)
	}
}

func (app *App) endpointGroup(group *gin.RouterGroup, config *app.Config) {
	if config.PluginEndpointEnabled != nil && *config.PluginEndpointEnabled {
		group.HEAD("/:hook_id/*path", app.Endpoint(config))
		group.POST("/:hook_id/*path", app.Endpoint(config))
		group.GET("/:hook_id/*path", app.Endpoint(config))
		group.PUT("/:hook_id/*path", app.Endpoint(config))
		group.DELETE("/:hook_id/*path", app.Endpoint(config))
		group.OPTIONS("/:hook_id/*path", app.Endpoint(config))
	}
}

func (appRef *App) awsLambdaTransactionGroup(group *gin.RouterGroup, config *app.Config) {
	if config.Platform == app.PLATFORM_SERVERLESS {
		appRef.awsTransactionHandler = transaction.NewAWSTransactionHandler(
			time.Duration(config.MaxServerlessTransactionTimeout) * time.Second,
		)
		group.POST(
			"/transaction",
			service.HandleAWSPluginTransaction(appRef.awsTransactionHandler),
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
	group.POST("/install/upload/package", controllers.UploadPlugin(config))
	group.POST("/install/upload/bundle", controllers.UploadBundle(config))
	group.POST("/install/identifiers", controllers.InstallPluginFromIdentifiers(config))
	group.POST("/install/upgrade", controllers.UpgradePlugin(config))
	group.GET("/install/tasks/:id", controllers.FetchPluginInstallationTask)
	group.POST("/install/tasks/delete_all", controllers.DeleteAllPluginInstallationTasks)
	group.POST("/install/tasks/:id/delete", controllers.DeletePluginInstallationTask)
	group.POST("/install/tasks/:id/delete/*identifier", controllers.DeletePluginInstallationItemFromTask)
	group.GET("/install/tasks", controllers.FetchPluginInstallationTasks)
	group.GET("/fetch/manifest", controllers.FetchPluginManifest)
	group.GET("/fetch/identifier", controllers.FetchPluginFromIdentifier)
	group.POST("/uninstall", controllers.UninstallPlugin)
	group.GET("/list", controllers.ListPlugins)
	group.POST("/installation/fetch/batch", controllers.BatchFetchPluginInstallationByIDs)
	group.POST("/installation/missing", controllers.FetchMissingPluginInstallations)
	group.GET("/models", controllers.ListModels)
	group.GET("/tools", controllers.ListTools)
	group.GET("/tool", controllers.GetTool)
	group.POST("/tools/check_existence", controllers.CheckToolExistence)
	group.GET("/agent_strategies", controllers.ListAgentStrategies)
	group.GET("/agent_strategy", controllers.GetAgentStrategy)
}

func (app *App) pluginAssetGroup(group *gin.RouterGroup) {
	group.GET("/:id", controllers.GetAsset)
}

func (app *App) pprofGroup(group *gin.RouterGroup, config *app.Config) {
	if config.PPROFEnabled {
		group.Use(CheckingKey(config.ServerKey))

		group.GET("/", controllers.PprofIndex)
		group.GET("/cmdline", controllers.PprofCmdline)
		group.GET("/profile", controllers.PprofProfile)
		group.GET("/symbol", controllers.PprofSymbol)
		group.GET("/trace", controllers.PprofTrace)
		group.GET("/goroutine", controllers.PprofGoroutine)
		group.GET("/heap", controllers.PprofHeap)
		group.GET("/allocs", controllers.PprofAllocs)
		group.GET("/block", controllers.PprofBlock)
		group.GET("/mutex", controllers.PprofMutex)
		group.GET("/threadcreate", controllers.PprofThreadcreate)
	}
}
