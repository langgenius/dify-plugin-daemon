package app

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	ServerPort uint16 `envconfig:"SERVER_PORT" validate:"required"`

	PluginInnerApiKey string `envconfig:"PLUGIN_INNER_API_KEY" validate:"required"`
	PluginInnerApiURL string `envconfig:"PLUGIN_INNER_API_URL" validate:"required"`

	PluginRemoteInstallingHost             string `envconfig:"PLUGIN_REMOTE_INSTALLING_HOST"`
	PluginRemoteInstallingPort             uint16 `envconfig:"PLUGIN_REMOTE_INSTALLING_PORT"`
	PluginRemoteInstallingEnabled          bool   `envconfig:"PLUGIN_REMOTE_INSTALLING_ENABLED"`
	PluginRemoteInstallingMaxConn          int    `envconfig:"PLUGIN_REMOTE_INSTALLING_MAX_CONN"`
	PluginRemoteInstallServerEventLoopNums int    `envconfig:"PLUGIN_REMOTE_INSTALL_SERVER_EVENT_LOOP_NUMS"`

	PluginWebhookEnabled bool `envconfig:"PLUGIN_WEBHOOK_ENABLED"`

	StoragePath        string `envconfig:"STORAGE_PLUGIN_PATH"  validate:"required"`
	ProcessCachingPath string `envconfig:"PROCESS_CACHING_PATH"  validate:"required"`

	Platform PlatformType `envconfig:"PLATFORM" validate:"required"`

	RoutinePoolSize int `envconfig:"ROUTINE_POOL_SIZE" validate:"required"`

	RedisHost string `envconfig:"REDIS_HOST" validate:"required"`
	RedisPort uint16 `envconfig:"REDIS_PORT" validate:"required"`
	RedisPass string `envconfig:"REDIS_PASS" validate:"required"`

	DBUsername string `envconfig:"DB_USERNAME" validate:"required"`
	DBPassword string `envconfig:"DB_PASSWORD" validate:"required"`
	DBHost     string `envconfig:"DB_HOST" validate:"required"`
	DBPort     uint16 `envconfig:"DB_PORT" validate:"required"`
	DBDatabase string `envconfig:"DB_DATABASE" validate:"required"`
	DBSslMode  string `envconfig:"DB_SSL_MODE" validate:"required,oneof=disable require"`

	LifetimeCollectionHeartbeatInterval int `envconfig:"LIFETIME_COLLECTION_HEARTBEAT_INTERVAL"  validate:"required"`
	LifetimeCollectionGCInterval        int `envconfig:"LIFETIME_COLLECTION_GC_INTERVAL" validate:"required"`
	LifetimeStateGCInterval             int `envconfig:"LIFETIME_STATE_GC_INTERVAL" validate:"required"`

	DifyInvocationConnectionIdleTimeout int `envconfig:"DIFY_INVOCATION_CONNECTION_IDLE_TIMEOUT" validate:"required"`
}

func (c *Config) Validate() error {
	validator := validator.New()
	err := validator.Struct(c)
	if err != nil {
		return err
	}

	if c.PluginRemoteInstallingEnabled {
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

	return nil
}

type PlatformType string

const (
	PLATFORM_LOCAL      PlatformType = "local"
	PLATFORM_AWS_LAMBDA PlatformType = "aws_lambda"
)
