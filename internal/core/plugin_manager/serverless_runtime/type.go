package serverless_runtime

import (
	"net/http"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type ServerlessPluginRuntime struct {
	basic_runtime.BasicChecksum
	plugin_entities.PluginRuntime

	// access url for the lambda function
	LambdaURL  string
	LambdaName string

	// listeners mapping session id to the listener
	listeners mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]

	client *http.Client

	PluginMaxExecutionTimeout int // in seconds

	RuntimeBufferSize    int
	RuntimeMaxBufferSize int
}

type ServerlessPluginRuntimeConfig struct {
	LambdaURL                 string
	LambdaName                string
	PluginMaxExecutionTimeout int
	RuntimeBufferSize         int
	RuntimeMaxBufferSize      int
}

func NewServerlessPluginRuntime(
	basicChecksum basic_runtime.BasicChecksum,
	pluginRuntime plugin_entities.PluginRuntime,
	config ServerlessPluginRuntimeConfig,
) *ServerlessPluginRuntime {
	// set default buffer sizes if not configured
	RuntimeBufferSize := config.RuntimeBufferSize
	if RuntimeBufferSize <= 0 {
		RuntimeBufferSize = 1024
	}
	RuntimeMaxBufferSize := config.RuntimeMaxBufferSize
	if RuntimeMaxBufferSize <= 0 {
		RuntimeMaxBufferSize = 5 * 1024 * 1024
	}

	return &ServerlessPluginRuntime{
		BasicChecksum:             basicChecksum,
		PluginRuntime:             pluginRuntime,
		LambdaURL:                 config.LambdaURL,
		LambdaName:                config.LambdaName,
		PluginMaxExecutionTimeout: config.PluginMaxExecutionTimeout,
		RuntimeBufferSize:         RuntimeBufferSize,
		RuntimeMaxBufferSize:      RuntimeMaxBufferSize,
	}
}
