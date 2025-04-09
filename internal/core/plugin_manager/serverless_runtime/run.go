package serverless_runtime

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

func (r *AWSPluginRuntime) StartPlugin() error {
	return nil
}

func (r *AWSPluginRuntime) Wait() (<-chan bool, error) {
	return nil, nil
}

func (r *AWSPluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS
}
