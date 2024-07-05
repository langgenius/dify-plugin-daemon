package app

type Config struct {
	DifyURL         string `envconfig:"DIFY_URL"`
	DifyCallingKey  string `envconfig:"DIFY_CALLING_KEY"`
	DifyCallingPort int16  `envconfig:"DIFY_CALLING_PORT"`

	PluginHost string `envconfig:"PLUGIN_HOST"`
	PluginPort int16  `envconfig:"PLUGIN_PORT"`

	StoragePath string `envconfig:"STORAGE_PATH"`

	Platform string `envconfig:"PLATFORM"`

	RoutinePoolSize int `envconfig:"ROUTINE_POOL_SIZE"`

	RedisHost string `envconfig:"REDIS_HOST"`
	RedisPort int16  `envconfig:"REDIS_PORT"`
	RedisPass string `envconfig:"REDIS_PASS"`

	LifetimeCollectionHeartbeatInterval int `envconfig:"LIFETIME_COLLECTION_HEARTBEAT_INTERVAL"`
	LifetimeCollectionGCInterval        int `envconfig:"LIFETIME_COLLECTION_GC_INTERVAL"`
	LifetimeStateGCInterval             int `envconfig:"LIFETIME_STATE_GC_INTERVAL"`
}

const (
	PLATFORM_LOCAL      = "local"
	PLATFORM_AWS_LAMBDA = "aws_lambda"
)
