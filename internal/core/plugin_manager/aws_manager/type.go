package aws_manager

import (
	"net/http"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
)

type AWSPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	plugin_entities.PluginRuntime

	// access url for the lambda function
	lambda_url  string
	lambda_name string

	// listeners mapping session id to the listener
	listeners mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]

	client *http.Client
}
