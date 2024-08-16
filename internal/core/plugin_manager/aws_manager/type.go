package aws_manager

import (
	"io"
	"net/http"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
)

type AWSPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	entities.PluginRuntime

	// access url for the lambda function
	lambda_url  string
	lambda_name string

	client *http.Client

	session_pool mapping.Map[string, *io.PipeWriter]

	// data stream take responsibility of listen to response from lambda and redirect to runtime listener
	data_stream chan []byte
}
