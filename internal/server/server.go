package server

import (
	"github.com/getsentry/sentry-go"
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/persistence"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/oss/local"
	"github.com/langgenius/dify-plugin-daemon/internal/oss/s3"
	"github.com/langgenius/dify-plugin-daemon/internal/oss/tencent_cos"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func initOSS(config *app.Config) oss.OSS {
	// init oss
	var oss oss.OSS
	var err error
	if config.PluginStorageType == "aws_s3" {
		oss, err = s3.NewS3Storage(
			config.S3UseAwsManagedIam,
			config.S3Endpoint,
			config.AWSAccessKey,
			config.AWSSecretKey,
			config.PluginStorageOSSBucket,
			config.AWSRegion,
		)
		if err != nil {
			log.Panic("Failed to create s3 storage: %s", err)
		}
	} else if config.PluginStorageType == "local" {
		oss = local.NewLocalStorage(config.PluginStorageLocalRoot)
	} else if config.PluginStorageType == "tencent_cos" {
		oss, err = tencent_cos.NewTencentCOSStorage(
			config.TencentCOSSecretId,
			config.TencentCOSSecretKey,
			config.TencentCOSRegion,
			config.PluginStorageOSSBucket)
		if err != nil {
			log.Panic("Failed to create tencent cos storage: %s", err)
		}
	} else {
		log.Panic("Invalid plugin storage type: %s", config.PluginStorageType)
	}

	return oss
}

func (app *App) Run(config *app.Config) {
	// init routine pool
	if config.SentryEnabled {
		routine.InitPool(config.RoutinePoolSize, sentry.ClientOptions{
			Dsn:              config.SentryDSN,
			AttachStacktrace: config.SentryAttachStacktrace,
			TracesSampleRate: config.SentryTracesSampleRate,
			SampleRate:       config.SentrySampleRate,
			EnableTracing:    config.SentryTracingEnabled,
		})
	} else {
		routine.InitPool(config.RoutinePoolSize)
	}

	// init db
	db.Init(config)

	// init oss
	oss := initOSS(config)

	// create manager
	manager := plugin_manager.InitGlobalManager(oss, config)

	// create cluster
	app.cluster = cluster.NewCluster(config, manager)

	// register plugin lifetime event
	manager.AddPluginRegisterHandler(app.cluster.RegisterPlugin)

	// init manager
	manager.Launch(config)

	// init persistence
	persistence.InitPersistence(oss, config)

	// launch cluster
	app.cluster.Launch()

	// start http server
	app.server(config)

	// block
	select {}
}
