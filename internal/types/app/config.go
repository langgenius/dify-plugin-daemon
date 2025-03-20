package app

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	// server
	ServerPort uint16 `envconfig:"SERVER_PORT" validate:"required"`
	ServerKey  string `envconfig:"SERVER_KEY" validate:"required"`

	// dify inner api
	DifyInnerApiURL string `envconfig:"DIFY_INNER_API_URL" validate:"required"`
	DifyInnerApiKey string `envconfig:"DIFY_INNER_API_KEY" validate:"required"`

	S3UseAwsManagedIam bool   `envconfig:"S3_USE_AWS_MANAGED_IAM" default:"true"`
	S3Endpoint         string `envconfig:"S3_ENDPOINT"`
	AWSAccessKey       string `envconfig:"AWS_ACCESS_KEY"`
	AWSSecretKey       string `envconfig:"AWS_SECRET_KEY"`
	AWSRegion          string `envconfig:"AWS_REGION"`

	TencentCOSSecretKey string `envconfig:"TENCENT_COS_SECRET_KEY"`
	TencentCOSSecretId  string `envconfig:"TENCENT_COS_SECRET_ID"`
	TencentCOSRegion    string `envconfig:"TENCENT_COS_REGION"`

	PluginStorageType      string `envconfig:"PLUGIN_STORAGE_TYPE" validate:"required,oneof=local aws_s3 tencent_cos"`
	PluginStorageOSSBucket string `envconfig:"PLUGIN_STORAGE_OSS_BUCKET"`
	PluginStorageLocalRoot string `envconfig:"PLUGIN_STORAGE_LOCAL_ROOT"`

	// plugin remote installing
	PluginRemoteInstallingHost                string `envconfig:"PLUGIN_REMOTE_INSTALLING_HOST"`
	PluginRemoteInstallingPort                uint16 `envconfig:"PLUGIN_REMOTE_INSTALLING_PORT"`
	PluginRemoteInstallingEnabled             *bool  `envconfig:"PLUGIN_REMOTE_INSTALLING_ENABLED"`
	PluginRemoteInstallingMaxConn             int    `envconfig:"PLUGIN_REMOTE_INSTALLING_MAX_CONN"`
	PluginRemoteInstallingMaxSingleTenantConn int    `envconfig:"PLUGIN_REMOTE_INSTALLING_MAX_SINGLE_TENANT_CONN"`
	PluginRemoteInstallServerEventLoopNums    int    `envconfig:"PLUGIN_REMOTE_INSTALL_SERVER_EVENT_LOOP_NUMS"`

	// plugin endpoint
	PluginEndpointEnabled *bool `envconfig:"PLUGIN_ENDPOINT_ENABLED"`

	// storage
	PluginWorkingPath      string `envconfig:"PLUGIN_WORKING_PATH"` // where the plugin finally running
	PluginMediaCacheSize   uint16 `envconfig:"PLUGIN_MEDIA_CACHE_SIZE"`
	PluginMediaCachePath   string `envconfig:"PLUGIN_MEDIA_CACHE_PATH"`
	PluginInstalledPath    string `envconfig:"PLUGIN_INSTALLED_PATH" validate:"required"` // where the plugin finally installed
	PluginPackageCachePath string `envconfig:"PLUGIN_PACKAGE_CACHE_PATH"`                 // where plugin packages stored

	// request timeout
	PluginMaxExecutionTimeout int `envconfig:"PLUGIN_MAX_EXECUTION_TIMEOUT" validate:"required"`

	// local launching max concurrent
	PluginLocalLaunchingConcurrent int `envconfig:"PLUGIN_LOCAL_LAUNCHING_CONCURRENT" validate:"required"`

	// platform like local or aws lambda
	Platform PlatformType `envconfig:"PLATFORM" validate:"required"`

	// routine pool
	RoutinePoolSize int `envconfig:"ROUTINE_POOL_SIZE" validate:"required"`

	// redis
	RedisHost   string `envconfig:"REDIS_HOST" validate:"required"`
	RedisPort   uint16 `envconfig:"REDIS_PORT" validate:"required"`
	RedisPass   string `envconfig:"REDIS_PASSWORD"`
	RedisUseSsl bool   `envconfig:"REDIS_USE_SSL"`

	// database
	DBUsername        string `envconfig:"DB_USERNAME" validate:"required"`
	DBPassword        string `envconfig:"DB_PASSWORD" validate:"required"`
	DBHost            string `envconfig:"DB_HOST" validate:"required"`
	DBPort            uint16 `envconfig:"DB_PORT" validate:"required"`
	DBDatabase        string `envconfig:"DB_DATABASE" validate:"required"`
	DBDefaultDatabase string `envconfig:"DB_DEFAULT_DATABASE" validate:"required"`
	DBSslMode         string `envconfig:"DB_SSL_MODE" validate:"required,oneof=disable require"`

	// persistence storage
	PersistenceStoragePath    string `envconfig:"PERSISTENCE_STORAGE_PATH"`
	PersistenceStorageMaxSize int64  `envconfig:"PERSISTENCE_STORAGE_MAX_SIZE"`

	// force verifying signature for all plugins, not allowing install plugin not signed
	ForceVerifyingSignature *bool `envconfig:"FORCE_VERIFYING_SIGNATURE"`

	// lifetime state management
	LifetimeCollectionHeartbeatInterval int `envconfig:"LIFETIME_COLLECTION_HEARTBEAT_INTERVAL"  validate:"required"`
	LifetimeCollectionGCInterval        int `envconfig:"LIFETIME_COLLECTION_GC_INTERVAL" validate:"required"`
	LifetimeStateGCInterval             int `envconfig:"LIFETIME_STATE_GC_INTERVAL" validate:"required"`

	DifyInvocationConnectionIdleTimeout int `envconfig:"DIFY_INVOCATION_CONNECTION_IDLE_TIMEOUT" validate:"required"`

	DifyPluginServerlessConnectorURL    *string `envconfig:"DIFY_PLUGIN_SERVERLESS_CONNECTOR_URL"`
	DifyPluginServerlessConnectorAPIKey *string `envconfig:"DIFY_PLUGIN_SERVERLESS_CONNECTOR_API_KEY"`

	MaxPluginPackageSize            int64 `envconfig:"MAX_PLUGIN_PACKAGE_SIZE" validate:"required"`
	MaxBundlePackageSize            int64 `envconfig:"MAX_BUNDLE_PACKAGE_SIZE" validate:"required"`
	MaxServerlessTransactionTimeout int   `envconfig:"MAX_SERVERLESS_TRANSACTION_TIMEOUT"`

	PythonInterpreterPath      string `envconfig:"PYTHON_INTERPRETER_PATH"`
	PythonEnvInitTimeout  	   int    `envconfig:"PYTHON_ENV_INIT_TIMEOUT" validate:"required"`
	PythonCompileAllExtraArgs  string `envconfig:"PYTHON_COMPILE_ALL_EXTRA_ARGS"`
	PipMirrorUrl               string `envconfig:"PIP_MIRROR_URL"`
	PipPreferBinary            *bool  `envconfig:"PIP_PREFER_BINARY"`
	PipVerbose                 *bool  `envconfig:"PIP_VERBOSE"`
	PipExtraArgs               string `envconfig:"PIP_EXTRA_ARGS"`

	DisplayClusterLog bool `envconfig:"DISPLAY_CLUSTER_LOG"`

	PPROFEnabled bool `envconfig:"PPROF_ENABLED"`

	SentryEnabled          bool    `envconfig:"SENTRY_ENABLED"`
	SentryDSN              string  `envconfig:"SENTRY_DSN"`
	SentryAttachStacktrace bool    `envconfig:"SENTRY_ATTACH_STACKTRACE"`
	SentryTracingEnabled   bool    `envconfig:"SENTRY_TRACING_ENABLED"`
	SentryTracesSampleRate float64 `envconfig:"SENTRY_TRACES_SAMPLE_RATE"`
	SentrySampleRate       float64 `envconfig:"SENTRY_SAMPLE_RATE"`

	// proxy settings
	HttpProxy  string `envconfig:"HTTP_PROXY"`
	HttpsProxy string `envconfig:"HTTPS_PROXY"`

	// log settings
	HealthApiLogEnabled *bool `envconfig:"HEALTH_API_LOG_ENABLED"`
}

func (c *Config) Validate() error {
	validator := validator.New()
	err := validator.Struct(c)
	if err != nil {
		return err
	}

	if c.PluginRemoteInstallingEnabled != nil && *c.PluginRemoteInstallingEnabled {
		if c.PluginRemoteInstallingHost == "" {
			return fmt.Errorf("plugin remote installing host is empty")
		}
		if c.PluginRemoteInstallingPort == 0 {
			return fmt.Errorf("plugin remote installing port is empty")
		}
		if c.PluginRemoteInstallingMaxConn == 0 {
			return fmt.Errorf("plugin remote installing max connection is empty")
		}
		if c.PluginRemoteInstallServerEventLoopNums == 0 {
			return fmt.Errorf("plugin remote install server event loop nums is empty")
		}
	}

	if c.Platform == PLATFORM_SERVERLESS {
		if c.DifyPluginServerlessConnectorURL == nil {
			return fmt.Errorf("dify plugin serverless connector url is empty")
		}

		if c.DifyPluginServerlessConnectorAPIKey == nil {
			return fmt.Errorf("dify plugin serverless connector api key is empty")
		}

		if c.MaxServerlessTransactionTimeout == 0 {
			return fmt.Errorf("max serverless transaction timeout is empty")
		}
	} else if c.Platform == PLATFORM_LOCAL {
		if c.PluginWorkingPath == "" {
			return fmt.Errorf("plugin working path is empty")
		}
	} else {
		return fmt.Errorf("invalid platform")
	}

	if c.PluginPackageCachePath == "" {
		return fmt.Errorf("plugin package cache path is empty")
	}

	if c.PluginStorageType == "aws_s3" {
		if c.PluginStorageOSSBucket == "" {
			return fmt.Errorf("plugin storage bucket is empty")
		}

		if c.AWSRegion == "" {
			return fmt.Errorf("aws region is empty")
		}
	}

	return nil
}

type PlatformType string

const (
	PLATFORM_LOCAL      PlatformType = "local"
	PLATFORM_SERVERLESS PlatformType = "serverless"
)
