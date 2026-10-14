package model_entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func chunkWithDelta(delta LLMResultChunkDelta) LLMResultChunk {
	return LLMResultChunk{Model: LLMModel("test-model"), Delta: delta}
}

func TestCarriesFirstToken(t *testing.T) {
	index := 0

	assert.False(t, chunkWithDelta(LLMResultChunkDelta{
		Index:   &index,
		Message: PromptMessage{Role: PROMPT_MESSAGE_ROLE_ASSISTANT, Content: ""},
	}).CarriesFirstToken())

	assert.True(t, chunkWithDelta(LLMResultChunkDelta{
		Index:   &index,
		Message: PromptMessage{Role: PROMPT_MESSAGE_ROLE_ASSISTANT, Content: "hi"},
	}).CarriesFirstToken())

	assert.False(t, chunkWithDelta(LLMResultChunkDelta{
		Index:   &index,
		Message: PromptMessage{Role: PROMPT_MESSAGE_ROLE_ASSISTANT, Content: []PromptMessageContent{}},
	}).CarriesFirstToken())

	assert.False(t, chunkWithDelta(LLMResultChunkDelta{
		Index: &index,
		Message: PromptMessage{
			Role:    PROMPT_MESSAGE_ROLE_ASSISTANT,
			Content: []PromptMessageContent{{Type: PROMPT_MESSAGE_CONTENT_TYPE_TEXT}},
		},
	}).CarriesFirstToken())

	assert.True(t, chunkWithDelta(LLMResultChunkDelta{
		Index: &index,
		Message: PromptMessage{
			Role:    PROMPT_MESSAGE_ROLE_ASSISTANT,
			Content: []PromptMessageContent{{Type: PROMPT_MESSAGE_CONTENT_TYPE_TEXT, Data: "hi"}},
		},
	}).CarriesFirstToken())

	assert.True(t, chunkWithDelta(LLMResultChunkDelta{
		Index: &index,
		Message: PromptMessage{
			Role:      PROMPT_MESSAGE_ROLE_ASSISTANT,
			Content:   "",
			ToolCalls: []PromptMessageToolCall{{ID: "call-1", Type: "function"}},
		},
	}).CarriesFirstToken())
}
