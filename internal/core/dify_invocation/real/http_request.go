package real

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func Request[T any](i *RealBackwardsInvocation, method string, path string, options ...http_requests.HttpOptions) (*T, error) {
	options = append(options,
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": i.PLUGIN_INNER_API_KEY,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)

	return http_requests.RequestAndParse[T](i.client, i.difyPath(path), method, options...)
}

func StreamResponse[T any](i *RealBackwardsInvocation, method string, path string, options ...http_requests.HttpOptions) (*stream.Stream[T], error) {
	options = append(
		options, http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": i.PLUGIN_INNER_API_KEY,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)

	return http_requests.RequestAndParseStream[T](i.client, i.difyPath(path), method, options...)
}

func (i *RealBackwardsInvocation) InvokeLLM(payload *dify_invocation.InvokeLLMRequest) (*stream.Stream[model_entities.LLMResultChunk], error) {
	return StreamResponse[model_entities.LLMResultChunk](i, "POST", "invoke/llm", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeTextEmbedding(payload *dify_invocation.InvokeTextEmbeddingRequest) (*model_entities.TextEmbeddingResult, error) {
	return Request[model_entities.TextEmbeddingResult](i, "POST", "invoke/text-embedding", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeRerank(payload *dify_invocation.InvokeRerankRequest) (*model_entities.RerankResult, error) {
	return Request[model_entities.RerankResult](i, "POST", "invoke/rerank", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeTTS(payload *dify_invocation.InvokeTTSRequest) (*stream.Stream[model_entities.TTSResult], error) {
	return StreamResponse[model_entities.TTSResult](i, "POST", "invoke/tts", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeSpeech2Text(payload *dify_invocation.InvokeSpeech2TextRequest) (*model_entities.Speech2TextResult, error) {
	return Request[model_entities.Speech2TextResult](i, "POST", "invoke/speech2text", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeModeration(payload *dify_invocation.InvokeModerationRequest) (*model_entities.ModerationResult, error) {
	return Request[model_entities.ModerationResult](i, "POST", "invoke/moderation", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeTool(payload *dify_invocation.InvokeToolRequest) (*stream.Stream[tool_entities.ToolResponseChunk], error) {
	return StreamResponse[tool_entities.ToolResponseChunk](i, "POST", "invoke/tool", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeApp(payload *dify_invocation.InvokeAppRequest) (*stream.Stream[map[string]any], error) {
	return StreamResponse[map[string]any](i, "POST", "invoke/app", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeEncrypt(payload *dify_invocation.InvokeEncryptRequest) (map[string]any, error) {
	if !payload.EncryptRequired(payload.Data) {
		return payload.Data, nil
	}

	data, err := Request[map[string]any](i, "POST", "invoke/encrypt", http_requests.HttpPayloadJson(payload))
	if err != nil {
		return nil, err
	}

	return *data, nil
}
