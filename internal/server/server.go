package server

import (
	"github.com/getsentry/sentry-go"
	"github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/persistence"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func initOSS(config *app.Config) oss.OSS {
	// init storage
	var storage oss.OSS
	var err error
	storage, err = factory.Load(config.PluginStorageType, oss.OSSArgs{
		Local: &oss.Local{
			Path: config.PluginStorageLocalRoot,
		},
		S3: &oss.S3{
			UseAws:       config.S3UseAWS,
			Endpoint:     config.S3Endpoint,
			UsePathStyle: config.S3UsePathStyle,
			AccessKey:    config.AWSAccessKey,
			SecretKey:    config.AWSSecretKey,
			Bucket:       config.PluginStorageOSSBucket,
			Region:       config.AWSRegion,
			UseIamRole:   config.S3UseAwsManagedIam,
		},
		TencentCOS: &oss.TencentCOS{
			Region:    config.TencentCOSRegion,
			SecretID:  config.TencentCOSSecretId,
			SecretKey: config.TencentCOSSecretKey,
			Bucket:    config.PluginStorageOSSBucket,
		},
		AzureBlob: &oss.AzureBlob{
			ConnectionString: config.AzureBlobStorageConnectionString,
			ContainerName:    config.AzureBlobStorageContainerName,
		},
		GoogleCloudStorage: &oss.GoogleCloudStorage{
			Bucket:         config.PluginStorageOSSBucket,
			CredentialsB64: config.GoogleCloudStorageCredentialsB64,
		},
		AliyunOSS: &oss.AliyunOSS{
			Region:      config.AliyunOSSRegion,
			Endpoint:    config.AliyunOSSEndpoint,
			AccessKey:   config.AliyunOSSAccessKeyID,
			SecretKey:   config.AliyunOSSAccessKeySecret,
			AuthVersion: config.AliyunOSSAuthVersion,
			Path:        config.AliyunOSSPath,
			Bucket:      config.PluginStorageOSSBucket,
		},
		HuaweiOBS: &oss.HuaweiOBS{
			AccessKey: config.HuaweiOBSAccessKey,
			SecretKey: config.HuaweiOBSSecretKey,
			Server:    config.HuaweiOBSServer,
			Bucket:    config.PluginStorageOSSBucket,
		},
		VolcengineTOS: &oss.VolcengineTOS{
			Region:    config.VolcengineTOSRegion,
			Endpoint:  config.VolcengineTOSEndpoint,
			AccessKey: config.VolcengineTOSAccessKey,
			SecretKey: config.VolcengineTOSSecretKey,
			Bucket:    config.PluginStorageOSSBucket,
		},
	})
	if err != nil {
		log.Panic("Failed to create storage: %s", err)
	}

	return storage
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

	curd.Init(config)

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
