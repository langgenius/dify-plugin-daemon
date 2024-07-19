package dify_invocation

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func Request[T any](method string, path string, options ...http_requests.HttpOptions) (*T, error) {
	options = append(options,
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": PLUGIN_INNER_API_KEY,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(60000),
	)

	return http_requests.RequestAndParse[T](client, difyPath(path), method, options...)
}

func StreamResponse[T any](method string, path string, options ...http_requests.HttpOptions) (*stream.StreamResponse[T], error) {
	options = append(
		options, http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": PLUGIN_INNER_API_KEY,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(60000),
	)

	return http_requests.RequestAndParseStream[T](client, difyPath(path), method, options...)
}

func InvokeModel(payload *InvokeModelRequest) (*stream.StreamResponse[InvokeModelResponseChunk], error) {
	return StreamResponse[InvokeModelResponseChunk]("POST", "invoke/model", http_requests.HttpPayloadJson(payload))
}

func InvokeTool(payload *InvokeToolRequest) (*stream.StreamResponse[InvokeToolResponseChunk], error) {
	return StreamResponse[InvokeToolResponseChunk]("POST", "invoke/tool", http_requests.HttpPayloadJson(payload))
}

func InvokeNode[T WorkflowNodeData](payload *InvokeNodeRequest[T]) (*InvokeNodeResponse, error) {
	return Request[InvokeNodeResponse]("POST", "invoke/node", http_requests.HttpPayloadJson(payload))
}
