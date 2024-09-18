package tester

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

type MockedDifyInvocation struct{}

func NewMockedDifyInvocation() dify_invocation.BackwardsInvocation {
	return &MockedDifyInvocation{}
}

func (m *MockedDifyInvocation) InvokeLLM(payload *dify_invocation.InvokeLLMRequest) (*stream.Stream[model_entities.LLMResultChunk], error) {
	stream := stream.NewStream[model_entities.LLMResultChunk](5)
	go func() {
		stream.Write(model_entities.LLMResultChunk{
			Model:             model_entities.LLMModel(payload.Model),
			PromptMessages:    payload.PromptMessages,
			SystemFingerprint: "test",
			Delta: model_entities.LLMResultChunkDelta{
				Index: &[]int{1}[0],
				Message: model_entities.PromptMessage{
					Role:    model_entities.PROMPT_MESSAGE_ROLE_ASSISTANT,
					Content: "hello",
					Name:    "test",
				},
			},
		})
		time.Sleep(100 * time.Millisecond)
		stream.Write(model_entities.LLMResultChunk{
			Model:             model_entities.LLMModel(payload.Model),
			PromptMessages:    payload.PromptMessages,
			SystemFingerprint: "test",
			Delta: model_entities.LLMResultChunkDelta{
				Index: &[]int{1}[0],
				Message: model_entities.PromptMessage{
					Role:    model_entities.PROMPT_MESSAGE_ROLE_ASSISTANT,
					Content: " world",
					Name:    "test",
				},
			},
		})
		time.Sleep(100 * time.Millisecond)
		stream.Write(model_entities.LLMResultChunk{
			Model:             model_entities.LLMModel(payload.Model),
			PromptMessages:    payload.PromptMessages,
			SystemFingerprint: "test",
			Delta: model_entities.LLMResultChunkDelta{
				Index: &[]int{2}[0],
				Message: model_entities.PromptMessage{
					Role:    model_entities.PROMPT_MESSAGE_ROLE_ASSISTANT,
					Content: " world",
					Name:    "test",
				},
			},
		})
		time.Sleep(100 * time.Millisecond)
		stream.Write(model_entities.LLMResultChunk{
			Model:             model_entities.LLMModel(payload.Model),
			PromptMessages:    payload.PromptMessages,
			SystemFingerprint: "test",
			Delta: model_entities.LLMResultChunkDelta{
				Index: &[]int{3}[0],
				Message: model_entities.PromptMessage{
					Role:    model_entities.PROMPT_MESSAGE_ROLE_ASSISTANT,
					Content: " !",
					Name:    "test",
				},
			},
		})
		time.Sleep(100 * time.Millisecond)
		stream.Write(model_entities.LLMResultChunk{
			Model:             model_entities.LLMModel(payload.Model),
			PromptMessages:    payload.PromptMessages,
			SystemFingerprint: "test",
			Delta: model_entities.LLMResultChunkDelta{
				Index: &[]int{3}[0],
				Usage: &model_entities.LLMUsage{
					PromptTokens:     &[]int{100}[0],
					CompletionTokens: &[]int{100}[0],
					TotalTokens:      &[]int{200}[0],
					Latency:          &[]float64{0.4}[0],
					Currency:         &[]string{"USD"}[0],
				},
			},
		})
		stream.Close()
	}()
	return stream, nil
}

func (m *MockedDifyInvocation) InvokeTextEmbedding(payload *dify_invocation.InvokeTextEmbeddingRequest) (*model_entities.TextEmbeddingResult, error) {
	result := model_entities.TextEmbeddingResult{
		Model: payload.Model,
		Usage: model_entities.EmbeddingUsage{
			Tokens:      &[]int{100}[0],
			TotalTokens: &[]int{100}[0],
			Latency:     &[]float64{0.1}[0],
			Currency:    &[]string{"USD"}[0],
		},
	}

	for range payload.Texts {
		result.Embeddings = append(result.Embeddings, []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0})
	}

	return &result, nil
}

func (m *MockedDifyInvocation) InvokeRerank(payload *dify_invocation.InvokeRerankRequest) (*model_entities.RerankResult, error) {
	result := model_entities.RerankResult{
		Model: payload.Model,
	}
	for i, doc := range payload.Docs {
		result.Docs = append(result.Docs, model_entities.RerankDocument{
			Index: &[]int{i}[0],
			Score: &[]float64{0.1}[0],
			Text:  &doc,
		})
	}
	return &result, nil
}

func (m *MockedDifyInvocation) InvokeTTS(payload *dify_invocation.InvokeTTSRequest) (*stream.Stream[model_entities.TTSResult], error) {
	stream := stream.NewStream[model_entities.TTSResult](5)
	go func() {
		for i := 0; i < 10; i++ {
			stream.Write(model_entities.TTSResult{
				Result: "a1b2c3d4",
			})
			time.Sleep(100 * time.Millisecond)
		}
		stream.Close()
	}()
	return stream, nil
}

func (m *MockedDifyInvocation) InvokeSpeech2Text(payload *dify_invocation.InvokeSpeech2TextRequest) (*model_entities.Speech2TextResult, error) {
	result := model_entities.Speech2TextResult{
		Result: "hello world",
	}
	return &result, nil
}

func (m *MockedDifyInvocation) InvokeModeration(payload *dify_invocation.InvokeModerationRequest) (*model_entities.ModerationResult, error) {
	result := model_entities.ModerationResult{
		Result: true,
	}
	return &result, nil
}

func (m *MockedDifyInvocation) InvokeTool(payload *dify_invocation.InvokeToolRequest) (*stream.Stream[tool_entities.ToolResponseChunk], error) {
	stream := stream.NewStream[tool_entities.ToolResponseChunk](5)
	go func() {
		for i := 0; i < 10; i++ {
			stream.Write(tool_entities.ToolResponseChunk{
				Type: tool_entities.ToolResponseChunkTypeText,
				Message: map[string]any{
					"text": "hello world",
				},
			})
			time.Sleep(100 * time.Millisecond)
		}
		stream.Close()
	}()

	return stream, nil
}

func (m *MockedDifyInvocation) InvokeApp(payload *dify_invocation.InvokeAppRequest) (*stream.Stream[map[string]any], error) {
	stream := stream.NewStream[map[string]any](5)
	go func() {
		for i := 0; i < 10; i++ {
			stream.Write(map[string]any{
				"text": "hello world",
			})
			time.Sleep(100 * time.Millisecond)
		}
		stream.Close()
	}()
	return stream, nil
}

func (m *MockedDifyInvocation) InvokeEncrypt(payload *dify_invocation.InvokeEncryptRequest) (map[string]any, error) {
	return payload.Data, nil
}
