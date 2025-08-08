package config

import (
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

// PluginInvoker defines a plugin invocation controller configuration
type PluginInvoker struct {
	// Name is the name of the controller function
	Name string
	// RequestType is the request type for this controller
	RequestType any
	// ResponseType is the response type for this controller
	ResponseType any
	// ResponseTypeName is the name of the response type
	ResponseTypeName string
	// BufferSize is the size of the response buffer
	BufferSize int
}

// PluginInvokers is a map of plugin invoker configurations
var PluginInvokers = map[string]PluginInvoker{
	"InvokeLLM": {
		Name:             "InvokeLLM",
		RequestType:      plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM]{},
		ResponseType:     model_entities.LLMResultChunk{},
		ResponseTypeName: "LLMResultChunk",
		BufferSize:       512,
	},
	"InvokeTextEmbedding": {
		Name:             "InvokeTextEmbedding",
		RequestType:      plugin_entities.InvokePluginRequest[requests.RequestInvokeTextEmbedding]{},
		ResponseType:     model_entities.TextEmbeddingResult{},
		ResponseTypeName: "TextEmbeddingResult",
		BufferSize:       1,
	},
	// ... other plugin invokers
}
