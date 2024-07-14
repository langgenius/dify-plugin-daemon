package app

type Config struct {
	SERVER_PORT int16 `envconfig:"SERVER_PORT"`

	PluginInnerApiKey string `envconfig:"PLUGIN_INNER_API_KEY"`
	PluginInnerApiURL string `envconfig:"PLUGIN_INNER_API_URL"`

	PluginRemoteInstallingHost string `envconfig:"PLUGIN_REMOTE_INSTALLING_HOST"`
	PluginRemoteInstallingPort int16  `envconfig:"PLUGIN_REMOTE_INSTALLING_PORT"`

	StoragePath string `envconfig:"STORAGE_PATH"`

	Platform PlatformType `envconfig:"PLATFORM"`

	RoutinePoolSize int `envconfig:"ROUTINE_POOL_SIZE"`

	RedisHost string `envconfig:"REDIS_HOST"`
	RedisPort int16  `envconfig:"REDIS_PORT"`
	RedisPass string `envconfig:"REDIS_PASS"`

	LifetimeCollectionHeartbeatInterval int `envconfig:"LIFETIME_COLLECTION_HEARTBEAT_INTERVAL"`
	LifetimeCollectionGCInterval        int `envconfig:"LIFETIME_COLLECTION_GC_INTERVAL"`
	LifetimeStateGCInterval             int `envconfig:"LIFETIME_STATE_GC_INTERVAL"`

	DifyInvocationConnectionIdleTimeout int `envconfig:"DIFY_INVOCATION_CONNECTION_IDLE_TIMEOUT"`
}

type PlatformType string

const (
	PLATFORM_LOCAL      PlatformType = "local"
	PLATFORM_AWS_LAMBDA PlatformType = "aws_lambda"
)
