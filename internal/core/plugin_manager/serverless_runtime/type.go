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

	stdoutBufferSize    int
	stdoutMaxBufferSize int
}

type ServerlessPluginRuntimeConfig struct {
	LambdaURL                 string
	LambdaName                string
	PluginMaxExecutionTimeout int
	StdoutBufferSize          int
	StdoutMaxBufferSize       int
}

func NewServerlessPluginRuntime(
	basicChecksum basic_runtime.BasicChecksum,
	pluginRuntime plugin_entities.PluginRuntime,
	config ServerlessPluginRuntimeConfig,
) *ServerlessPluginRuntime {
	// set default buffer sizes if not configured
	stdoutBufferSize := config.StdoutBufferSize
	if stdoutBufferSize <= 0 {
		stdoutBufferSize = 1024
	}
	stdoutMaxBufferSize := config.StdoutMaxBufferSize
	if stdoutMaxBufferSize <= 0 {
		stdoutMaxBufferSize = 5 * 1024 * 1024
	}

	return &ServerlessPluginRuntime{
		BasicChecksum:             basicChecksum,
		PluginRuntime:             pluginRuntime,
		LambdaURL:                 config.LambdaURL,
		LambdaName:                config.LambdaName,
		PluginMaxExecutionTimeout: config.PluginMaxExecutionTimeout,
		stdoutBufferSize:          stdoutBufferSize,
		stdoutMaxBufferSize:       stdoutMaxBufferSize,
	}
}
