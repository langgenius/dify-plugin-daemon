package real

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// Send a request to dify inner api and validate the response
func Request[T any](i *RealBackwardsInvocation, method string, path string, options ...http_requests.HttpOptions) (*T, error) {
	options = append(options,
		http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": i.dify_inner_api_key,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)

	req, err := http_requests.RequestAndParse[BaseBackwardsInvocationResponse[T]](i.client, i.difyPath(path), method, options...)
	if err != nil {
		return nil, err
	}

	if req.Error != "" {
		return nil, fmt.Errorf("request failed: %s", req.Error)
	}

	if req.Data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	if err := validators.GlobalEntitiesValidator.Struct(req.Data); err != nil {
		return nil, fmt.Errorf("validate request failed: %s", err.Error())
	}

	return req.Data, nil
}

func StreamResponse[T any](i *RealBackwardsInvocation, method string, path string, options ...http_requests.HttpOptions) (
	*stream.Stream[T], error,
) {
	options = append(
		options, http_requests.HttpHeader(map[string]string{
			"X-Inner-Api-Key": i.dify_inner_api_key,
		}),
		http_requests.HttpWriteTimeout(5000),
		http_requests.HttpReadTimeout(240000),
	)

	response, err := http_requests.RequestAndParseStream[BaseBackwardsInvocationResponse[T]](i.client, i.difyPath(path), method, options...)
	if err != nil {
		return nil, err
	}

	new_response := stream.NewStream[T](1024)
	new_response.OnClose(func() {
		response.Close()
	})
	routine.Submit(func() {
		defer new_response.Close()
		for response.Next() {
			t, err := response.Read()
			if err != nil {
				new_response.WriteError(err)
				break
			}

			if t.Error != "" {
				new_response.WriteError(fmt.Errorf("request failed: %s", t.Error))
				break
			}

			if t.Data == nil {
				new_response.WriteError(fmt.Errorf("data is nil"))
				break
			}

			if err := validators.GlobalEntitiesValidator.Struct(t.Data); err != nil {
				new_response.WriteError(fmt.Errorf("validate request failed: %s", err.Error()))
				break
			}

			new_response.Write(*t.Data)
		}
	})

	return new_response, nil
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

func (i *RealBackwardsInvocation) InvokeParameterExtractor(payload *dify_invocation.InvokeParameterExtractorRequest) (*dify_invocation.InvokeNodeResponse, error) {
	return Request[dify_invocation.InvokeNodeResponse](i, "POST", "invoke/parameter-extractor", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeQuestionClassifier(payload *dify_invocation.InvokeQuestionClassifierRequest) (*dify_invocation.InvokeNodeResponse, error) {
	return Request[dify_invocation.InvokeNodeResponse](i, "POST", "invoke/question-classifier", http_requests.HttpPayloadJson(payload))
}

func (i *RealBackwardsInvocation) InvokeEncrypt(payload *dify_invocation.InvokeEncryptRequest) (map[string]any, error) {
	if !payload.EncryptRequired(payload.Data) {
		return payload.Data, nil
	}

	type resp struct {
		Data map[string]any `json:"data,omitempty"`
	}

	data, err := Request[resp](i, "POST", "invoke/encrypt", http_requests.HttpPayloadJson(payload))
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}

func (i *RealBackwardsInvocation) InvokeSummary(payload *dify_invocation.InvokeSummaryRequest) (*dify_invocation.InvokeSummaryResponse, error) {
	return Request[dify_invocation.InvokeSummaryResponse](i, "POST", "invoke/summary", http_requests.HttpPayloadJson(payload))
}
