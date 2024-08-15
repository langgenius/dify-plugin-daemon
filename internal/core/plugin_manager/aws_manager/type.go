package aws_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

type AWSPluginRuntime struct {
	positive_manager.PositivePluginRuntime
	entities.PluginRuntime

	// access url for the lambda function
	lambda_url  string
	lambda_name string
}
