package aws_manager

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"

func (r *AWSPluginRuntime) StartPlugin() error {
	return nil
}

func (r *AWSPluginRuntime) Wait() (<-chan bool, error) {
	return nil, nil
}

func (r *AWSPluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_AWS
}
