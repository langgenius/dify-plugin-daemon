package aws_manager

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities"

func (r *AWSPluginRuntime) StartPlugin() error {
	return nil
}

func (r *AWSPluginRuntime) Wait() (<-chan bool, error) {
	return nil, nil
}

func (r *AWSPluginRuntime) Type() entities.PluginRuntimeType {
	return entities.PLUGIN_RUNTIME_TYPE_AWS
}
