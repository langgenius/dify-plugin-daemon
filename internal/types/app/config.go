package app

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	// server
	ServerPort uint16 `envconfig:"SERVER_PORT" validate:"required"`
	ServerKey  string `envconfig:"SERVER_KEY" validate:"required"`

	// admin api enable
	AdminApiEnabled bool   `envconfig:"ADMIN_API_ENABLED" default:"false"`
	AdminApiKey     string `envconfig:"ADMIN_API_KEY"`

	// dify inner api
	DifyInnerApiURL string `envconfig:"DIFY_INNER_API_URL" validate:"required"`
	DifyInnerApiKey string `envconfig:"DIFY_INNER_API_KEY" validate:"required"`

	// storage config
	// https://github.com/langgenius/dify-cloud-kit/blob/main/oss/factory/factory.go
	PluginStorageType      string `envconfig:"PLUGIN_STORAGE_TYPE" validate:"required"`
	PluginStorageOSSBucket string `envconfig:"PLUGIN_STORAGE_OSS_BUCKET"`

	// aws s3
	S3UseAwsManagedIam bool   `envconfig:"S3_USE_AWS_MANAGED_IAM" default:"false"`
	S3UseAWS           bool   `envconfig:"S3_USE_AWS" default:"true"`
	S3Endpoint         string `envconfig:"S3_ENDPOINT"`
	S3UsePathStyle     bool   `envconfig:"S3_USE_PATH_STYLE" default:"true"`
	AWSAccessKey       string `envconfig:"AWS_ACCESS_KEY"`
	AWSSecretKey       string `envconfig:"AWS_SECRET_KEY"`
	AWSRegion          string `envconfig:"AWS_REGION"`

	// tencent cos
	TencentCOSSecretKey string `envconfig:"TENCENT_COS_SECRET_KEY"`
	TencentCOSSecretId  string `envconfig:"TENCENT_COS_SECRET_ID"`
	TencentCOSRegion    string `envconfig:"TENCENT_COS_REGION"`

	// azure blob
	AzureBlobStorageContainerName    string `envconfig:"AZURE_BLOB_STORAGE_CONTAINER_NAME"`
	AzureBlobStorageConnectionString string `envconfig:"AZURE_BLOB_STORAGE_CONNECTION_STRING"`

	// aliyun oss
	AliyunOSSRegion          string `envconfig:"ALIYUN_OSS_REGION"`
	AliyunOSSEndpoint        string `envconfig:"ALIYUN_OSS_ENDPOINT"`
	AliyunOSSAccessKeyID     string `envconfig:"ALIYUN_OSS_ACCESS_KEY_ID"`
	AliyunOSSAccessKeySecret string `envconfig:"ALIYUN_OSS_ACCESS_KEY_SECRET"`
	AliyunOSSAuthVersion     string `envconfig:"ALIYUN_OSS_AUTH_VERSION" default:"v4"`
	AliyunOSSPath            string `envconfig:"ALIYUN_OSS_PATH"`

	// google gcs
	GoogleCloudStorageCredentialsB64 string `envconfig:"GCS_CREDENTIALS"`

	// huawei obs
	HuaweiOBSAccessKey string `envconfig:"HUAWEI_OBS_ACCESS_KEY"`
	HuaweiOBSSecretKey string `envconfig:"HUAWEI_OBS_SECRET_KEY"`
	HuaweiOBSServer    string `envconfig:"HUAWEI_OBS_SERVER"`

	// volcengine tos
	VolcengineTOSEndpoint  string `envconfig:"VOLCENGINE_TOS_ENDPOINT"`
	VolcengineTOSAccessKey string `envconfig:"VOLCENGINE_TOS_ACCESS_KEY"`
	VolcengineTOSSecretKey string `envconfig:"VOLCENGINE_TOS_SECRET_KEY"`
	VolcengineTOSRegion    string `envconfig:"VOLCENGINE_TOS_REGION"`

	// local
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

	// add a global reference to plugins to prevent them from being garbage collected
	// not allowed for local mode
	PluginAllowOrphans bool `envconfig:"PLUGIN_ALLOW_ORPHANS" default:"false"`

	// platform like local or aws lambda
	Platform PlatformType `envconfig:"PLATFORM" validate:"required"`

	// routine pool
	RoutinePoolSize int `envconfig:"ROUTINE_POOL_SIZE" validate:"required"`

	// redis
	RedisHost   string `envconfig:"REDIS_HOST"`
	RedisPort   uint16 `envconfig:"REDIS_PORT"`
	RedisPass   string `envconfig:"REDIS_PASSWORD"`
	RedisUser   string `envconfig:"REDIS_USERNAME"`
	RedisUseSsl bool   `envconfig:"REDIS_USE_SSL"`
	RedisDB     int    `envconfig:"REDIS_DB"`

	// redis sentinel
	RedisUseSentinel           bool    `envconfig:"REDIS_USE_SENTINEL"`
	RedisSentinels             string  `envconfig:"REDIS_SENTINELS"`
	RedisSentinelServiceName   string  `envconfig:"REDIS_SENTINEL_SERVICE_NAME"`
	RedisSentinelUsername      string  `envconfig:"REDIS_SENTINEL_USERNAME"`
	RedisSentinelPassword      string  `envconfig:"REDIS_SENTINEL_PASSWORD"`
	RedisSentinelSocketTimeout float64 `envconfig:"REDIS_SENTINEL_SOCKET_TIMEOUT"`

	// database
	DBType            string `envconfig:"DB_TYPE" default:"postgresql"`
	DBUsername        string `envconfig:"DB_USERNAME" validate:"required"`
	DBPassword        string `envconfig:"DB_PASSWORD" validate:"required"`
	DBHost            string `envconfig:"DB_HOST" validate:"required"`
	DBPort            uint16 `envconfig:"DB_PORT" validate:"required"`
	DBDatabase        string `envconfig:"DB_DATABASE" validate:"required"`
	DBDefaultDatabase string `envconfig:"DB_DEFAULT_DATABASE" validate:"required"`
	DBSslMode         string `envconfig:"DB_SSL_MODE" validate:"required,oneof=disable require"`

	// database connection pool settings
	DBMaxIdleConns    int    `envconfig:"DB_MAX_IDLE_CONNS" default:"10"`
	DBMaxOpenConns    int    `envconfig:"DB_MAX_OPEN_CONNS" default:"30"`
	DBConnMaxLifetime int    `envconfig:"DB_CONN_MAX_LIFETIME" default:"3600"`
	DBExtras          string `envconfig:"DB_EXTRAS"`
	DBCharset         string `envconfig:"DB_CHARSET"`

	// persistence storage
	PersistenceStoragePath    string `envconfig:"PERSISTENCE_STORAGE_PATH"`
	PersistenceStorageMaxSize int64  `envconfig:"PERSISTENCE_STORAGE_MAX_SIZE"`

	// force verifying signature for all plugins, not allowing install plugin not signed
	ForceVerifyingSignature *bool `envconfig:"FORCE_VERIFYING_SIGNATURE"`

	// enable or disable third-party signature verification for plugins
	ThirdPartySignatureVerificationEnabled bool `envconfig:"THIRD_PARTY_SIGNATURE_VERIFICATION_ENABLED"  default:"false"`
	// a comma-separated list of file paths to public keys in addition to the official public key for signature verification
	ThirdPartySignatureVerificationPublicKeys []string `envconfig:"THIRD_PARTY_SIGNATURE_VERIFICATION_PUBLIC_KEYS"  default:""`

	// Enforce signature verification for plugins claiming Langgenius authorship
	EnforceLanggeniusSignatures bool `envconfig:"ENFORCE_LANGGENIUS_PLUGIN_SIGNATURES" default:"true"`

	// lifetime state management
	LifetimeCollectionHeartbeatInterval int `envconfig:"LIFETIME_COLLECTION_HEARTBEAT_INTERVAL"  validate:"required"`
	LifetimeCollectionGCInterval        int `envconfig:"LIFETIME_COLLECTION_GC_INTERVAL" validate:"required"`
	LifetimeStateGCInterval             int `envconfig:"LIFETIME_STATE_GC_INTERVAL" validate:"required"`

	DifyInvocationConnectionIdleTimeout int `envconfig:"DIFY_INVOCATION_CONNECTION_IDLE_TIMEOUT" validate:"required"`

	DifyPluginServerlessConnectorURL           *string `envconfig:"DIFY_PLUGIN_SERVERLESS_CONNECTOR_URL"`
	DifyPluginServerlessConnectorAPIKey        *string `envconfig:"DIFY_PLUGIN_SERVERLESS_CONNECTOR_API_KEY"`
	DifyPluginServerlessConnectorLaunchTimeout int     `envconfig:"DIFY_PLUGIN_SERVERLESS_CONNECTOR_LAUNCH_TIMEOUT"`

	MaxPluginPackageSize            int64 `envconfig:"MAX_PLUGIN_PACKAGE_SIZE" validate:"required"`
	MaxBundlePackageSize            int64 `envconfig:"MAX_BUNDLE_PACKAGE_SIZE" validate:"required"`
	MaxServerlessTransactionTimeout int   `envconfig:"MAX_SERVERLESS_TRANSACTION_TIMEOUT"`

	PythonInterpreterPath     string `envconfig:"PYTHON_INTERPRETER_PATH"`
	UvPath                    string `envconfig:"UV_PATH"  default:""`
	PythonEnvInitTimeout      int    `envconfig:"PYTHON_ENV_INIT_TIMEOUT" validate:"required"`
	PythonCompileAllExtraArgs string `envconfig:"PYTHON_COMPILE_ALL_EXTRA_ARGS"`
	PipMirrorUrl              string `envconfig:"PIP_MIRROR_URL"`
	PipPreferBinary           *bool  `envconfig:"PIP_PREFER_BINARY"`
	PipVerbose                *bool  `envconfig:"PIP_VERBOSE"`
	PipExtraArgs              string `envconfig:"PIP_EXTRA_ARGS"`

	// Runtime buffer configuration (applies to both local and serverless runtimes)
	// These are the new generic names that should be used going forward
	PluginRuntimeBufferSize    int `envconfig:"PLUGIN_RUNTIME_BUFFER_SIZE" default:"1024"`
	PluginRuntimeMaxBufferSize int `envconfig:"PLUGIN_RUNTIME_MAX_BUFFER_SIZE" default:"5242880"`

	// Legacy STDIO-specific buffer configuration (kept for backward compatibility)
	// If the new PluginRuntime* configs are not set, these will be used as fallback
	PluginStdioBufferSize    int `envconfig:"PLUGIN_STDIO_BUFFER_SIZE" default:"1024"`
	PluginStdioMaxBufferSize int `envconfig:"PLUGIN_STDIO_MAX_BUFFER_SIZE" default:"5242880"`

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
	NoProxy    string `envconfig:"NO_PROXY"`

	// log settings
	HealthApiLogEnabled *bool `envconfig:"HEALTH_API_LOG_ENABLED"`

	// dify invocation write timeout in milliseconds
	DifyInvocationWriteTimeout int64 `envconfig:"DIFY_BACKWARDS_INVOCATION_WRITE_TIMEOUT" default:"5000"`
	// dify invocation read timeout in milliseconds
	DifyInvocationReadTimeout int64 `envconfig:"DIFY_BACKWARDS_INVOCATION_READ_TIMEOUT" default:"240000"`
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
		if c.PluginAllowOrphans {
			return fmt.Errorf("orphan plugins are not allowed in local mode")
		}
	} else {
		return fmt.Errorf("invalid platform")
	}

	if c.PluginPackageCachePath == "" {
		return fmt.Errorf("plugin package cache path is empty")
	}

	return nil
}

// Prefers Stdio (legacy) config if user has customized it, falls back to Runtime (new) config.
func (c *Config) GetLocalRuntimeBufferSize() int {
	if c.PluginStdioBufferSize != 1024 {
		return c.PluginStdioBufferSize
	}
	return c.PluginRuntimeBufferSize
}

// Prefers Stdio (legacy) config if user has customized it, falls back to Runtime (new) config.
func (c *Config) GetLocalRuntimeMaxBufferSize() int {
	if c.PluginStdioMaxBufferSize != 5242880 {
		return c.PluginStdioMaxBufferSize
	}
	return c.PluginRuntimeMaxBufferSize
}

type PlatformType string

const (
	PLATFORM_LOCAL      PlatformType = "local"
	PLATFORM_SERVERLESS PlatformType = "serverless"
)
