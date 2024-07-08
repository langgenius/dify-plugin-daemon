package dify_invocation

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func Request[T any](method string, path string, options ...requests.HttpOptions) (*T, error) {
	options = append(options, requests.HttpHeader(map[string]string{
		"X-Inner-Api-Key": PLUGIN_INNER_API_KEY,
	}))

	return requests.RequestAndParse[T](client, difyPath(path), method, options...)
}

func StreamResponse[T any](method string, path string, options ...requests.HttpOptions) (*stream.StreamResponse[T], error) {
	options = append(options, requests.HttpHeader(map[string]string{
		"X-Inner-Api-Key": PLUGIN_INNER_API_KEY,
	}))

	return requests.RequestAndParseStream[T](client, difyPath(path), method, options...)
}

func InvokeModel(payload InvokeModelRequest) (*stream.StreamResponse[InvokeModelResponseChunk], error) {
	return StreamResponse[InvokeModelResponseChunk]("POST", "invoke/model", requests.HttpPayloadJson(payload))
}

func InvokeTool(payload InvokeToolRequest) (*stream.StreamResponse[InvokeToolResponseChunk], error) {
	return StreamResponse[InvokeToolResponseChunk]("POST", "invoke/tool", requests.HttpPayloadJson(payload))
}

func InvokeNode[T WorkflowNodeData](payload InvokeNodeRequest[T]) (*InvokeNodeResponse, error) {
	return Request[InvokeNodeResponse]("POST", "invoke/node", requests.HttpPayloadJson(payload))
}
