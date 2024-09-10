package dify_invocation

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
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

func StreamResponse[T any](method string, path string, options ...http_requests.HttpOptions) (*stream.Stream[T], error) {
	options = append(
		options, http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": PLUGIN_INNER_API_KEY,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(60000),
	)

	return http_requests.RequestAndParseStream[T](client, difyPath(path), method, options...)
}

func InvokeLLM(payload *InvokeLLMRequest) (*stream.Stream[model_entities.LLMResultChunk], error) {
	return StreamResponse[model_entities.LLMResultChunk]("POST", "invoke/llm", http_requests.HttpPayloadJson(payload))
}

func InvokeTextEmbedding(payload *InvokeTextEmbeddingRequest) (*model_entities.TextEmbeddingResult, error) {
	return Request[model_entities.TextEmbeddingResult]("POST", "invoke/text-embedding", http_requests.HttpPayloadJson(payload))
}

func InvokeRerank(payload *InvokeRerankRequest) (*model_entities.RerankResult, error) {
	return Request[model_entities.RerankResult]("POST", "invoke/rerank", http_requests.HttpPayloadJson(payload))
}

func InvokeTTS(payload *InvokeTTSRequest) (*stream.Stream[model_entities.TTSResult], error) {
	return StreamResponse[model_entities.TTSResult]("POST", "invoke/tts", http_requests.HttpPayloadJson(payload))
}

func InvokeSpeech2Text(payload *InvokeSpeech2TextRequest) (*model_entities.Speech2TextResult, error) {
	return Request[model_entities.Speech2TextResult]("POST", "invoke/speech2text", http_requests.HttpPayloadJson(payload))
}

func InvokeModeration(payload *InvokeModerationRequest) (*model_entities.ModerationResult, error) {
	return Request[model_entities.ModerationResult]("POST", "invoke/moderation", http_requests.HttpPayloadJson(payload))
}

func InvokeTool(payload *InvokeToolRequest) (*stream.Stream[tool_entities.ToolResponseChunk], error) {
	return StreamResponse[tool_entities.ToolResponseChunk]("POST", "invoke/tool", http_requests.HttpPayloadJson(payload))
}

func InvokeApp(payload *InvokeAppRequest) (*stream.Stream[map[string]any], error) {
	return StreamResponse[map[string]any]("POST", "invoke/app", http_requests.HttpPayloadJson(payload))
}

func InvokeEncrypt(payload *InvokeEncryptRequest) (map[string]any, error) {
	if !payload.EncryptRequired(payload.Data) {
		return payload.Data, nil
	}

	data, err := Request[map[string]any]("POST", "invoke/encrypt", http_requests.HttpPayloadJson(payload))
	if err != nil {
		return nil, err
	}

	return *data, nil
}
