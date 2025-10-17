package model_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/stretchr/testify/assert"
)

func TestFullFunctionPromptMessage(t *testing.T) {
	const (
		system_message = `
		{
			"role": "system",
			"content": "you are a helpful assistant"
		}
		`
		user_message = `
		{
			"role": "user",
			"content": "hello"
		}`
		assistant_message = `
		{
			"role": "assistant",
			"content": "you are a helpful assistant"
		}`
		image_message = `
		{
			"role": "user",
			"content": [
				{
					"type": "image",
					"data": "base64"
				}
			]
		}`
		tool_message = `
		{
			"role": "tool",
			"content": "you are a helpful assistant",
			"tool_call_id": "123"
		}
		`
	)

	promptMessage, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(system_message))
	if err != nil {
		t.Error(err)
	}
	if promptMessage.Role != "system" {
		t.Error("role is not system")
	}

	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(user_message))
	if err != nil {
		t.Error(err)
	}
	if promptMessage.Role != "user" {
		t.Error("role is not user")
	}

	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(assistant_message))
	if err != nil {
		t.Error(err)
	}
	if promptMessage.Role != "assistant" {
		t.Error("role is not assistant")
	}

	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(image_message))
	if err != nil {
		t.Error(err)
	}
	if promptMessage.Role != "user" {
		t.Error("role is not user")
	}
	if promptMessage.Content.([]PromptMessageContent)[0].Type != "image" {
		t.Error("type is not image")
	}

	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(tool_message))
	if err != nil {
		t.Error(err)
	}
	if promptMessage.Role != "tool" {
		t.Error("role is not tool")
	}
	if promptMessage.ToolCallId != "123" {
		t.Error("tool call id is not 123")
	}
}

func TestWrongRole(t *testing.T) {
	const (
		wrong_role = `
		{
			"role": "wrong",
			"content": "you are a helpful assistant"
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(wrong_role))
	if err == nil {
		t.Error("error is nil")
	}
}

func TestWrongContent(t *testing.T) {
	const (
		wrong_content = `
		{
			"role": "user",
			"content": 123
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(wrong_content))
	if err == nil {
		t.Error("error is nil")
	}
}

func TestWrongContentArray(t *testing.T) {
	const (
		wrong_content_array = `
		{
			"role": "user",
			"content": [
				{
					"type": "image",
					"data": 123
				}
			]
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(wrong_content_array))
	if err == nil {
		t.Error("error is nil")
	}
}

// func TestWrongContentArray2(t *testing.T) {
// 	const (
// 		wrong_content_array2 = `
// 		{
// 			"role": "user",
// 			"content": [
// 				{
// 					"type": "image"
// 				}
// 			]
// 		}
// 		`
// 	)

// 	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(wrong_content_array2))
// 	if err == nil {
// 		t.Error("error is nil")
// 	}
// }

func TestWrongContentArray3(t *testing.T) {
	const (
		wrong_content_array3 = `
		{
			"role": "user",
			"content": [
				{
					"type": "wwww",
					"data": "base64"
				},
				{
					"type": "image",
					"data": "base64"
				}
			]
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(wrong_content_array3))
	if err == nil {
		t.Error("error is nil")
	}
}

func TestFullFunctionLLMResultChunk(t *testing.T) {
	const (
		llm_result_chunk = `
		{
			"model": "gpt-3.5-turbo",
			"prompt_messages": [
				{
					"role": "system",
					"content": "you are a helpful assistant"
				},
				{
					"role": "user",
					"content": "hello"
				}
			],
			"system_fingerprint": "123",
			"delta": {
				"index" : 0,
				"message" : {
					"role": "assistant",
					"content": "I am a helpful assistant"
				},
				"usage": {
					"prompt_tokens": 10,
					"prompt_unit_price": 0.1,
					"prompt_price_unit": 1,
					"prompt_price": 1,
					"completion_tokens": 10,
					"completion_unit_price": 0.1,
					"completion_price_unit": 1,
					"completion_price": 1,
					"total_tokens": 20,
					"total_price": 2,
					"currency": "usd",
					"latency": 0.1
				},
				"finish_reason": "completed"
			}
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[LLMResultChunk]([]byte(llm_result_chunk))
	if err != nil {
		t.Error(err)
	}
}

func TestZeroLLMUsage(t *testing.T) {
	const (
		llm_usage = `
		{
			"prompt_tokens": 0,
			"prompt_unit_price": 0,
			"prompt_price_unit": 0,
			"prompt_price": 0,
			"completion_tokens": 0,
			"completion_unit_price": 0,
			"completion_price_unit": 0,
			"completion_price": 0,
			"total_tokens": 0,
			"total_price": 0,
			"currency": "usd",
			"latency": 0
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[LLMUsage]([]byte(llm_usage))
	if err != nil {
		t.Error(err)
	}
}

func TestTextPromptMessage(t *testing.T) {
	const (
		promptMessage = `
		{
			"role": "user",
			"content": "hello"
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(promptMessage))
	if err != nil {
		t.Error(err)
	}
}

func TestImagePromptMessage(t *testing.T) {
	const (
		promptMessage = `
		{
			"role": "user",
			"content": [
				{
					"type": "image",
					"data": "base64"
				}
			]
		}
		`
	)

	_, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(promptMessage))
	if err != nil {
		t.Error(err)
	}
}

func TestLLMResultChunkCompatibility(t *testing.T) {
	llmResultChunk := LLMResultChunk{
		Model: "gpt-3.5-turbo",
	}

	result := parser.MarshalJson(llmResultChunk)
	assert.Contains(t, string(result), `"prompt_messages":[]`)

	llmResultChunkPointer := &llmResultChunk
	result = parser.MarshalJson(llmResultChunkPointer)
	assert.Contains(t, string(result), `"prompt_messages":[]`)
}

func TestMultiModalPromptMessageWithFilename(t *testing.T) {
	const (
		imageMessageWithFilename = `
		{
			"role": "user",
			"content": [
				{
					"type": "image",
					"data": "base64encodeddata",
					"filename": "example_image.jpg"
				}
			]
		}`

		documentMessageWithFilename = `
		{
			"role": "user",
			"content": [
				{
					"type": "document",
					"url": "https://example.com/document.pdf",
					"filename": "document.pdf",
					"mime_type": "application/pdf"
				}
			]
		}`

		audioMessageWithFilename = `
		{
			"role": "user",
			"content": [
				{
					"type": "audio",
					"base64_data": "base64audioencodeddata",
					"filename": "audio_sample.mp3",
					"format": "mp3"
				}
			]
		}`

		videoMessageWithFilename = `
		{
			"role": "user",
			"content": [
				{
					"type": "video",
					"url": "https://example.com/video.mp4",
					"filename": "video_sample.mp4"
				}
			]
		}`

		mixedContentWithFilename = `
		{
			"role": "user",
			"content": [
				{
					"type": "text",
					"data": "Please analyze this image"
				},
				{
					"type": "image",
					"data": "base64encodeddata",
					"filename": "screenshot.png"
				}
			]
		}`
	)

	// Test image message with filename
	promptMessage, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(imageMessageWithFilename))
	if err != nil {
		t.Error(err)
	}
	if promptMessage.Role != "user" {
		t.Error("role is not user")
	}
	content := promptMessage.Content.([]PromptMessageContent)
	if content[0].Type != "image" {
		t.Error("type is not image")
	}
	if content[0].Filename != "example_image.jpg" {
		t.Errorf("expected filename 'example_image.jpg', got '%s'", content[0].Filename)
	}

	// Test document message with filename
	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(documentMessageWithFilename))
	if err != nil {
		t.Error(err)
	}
	content = promptMessage.Content.([]PromptMessageContent)
	if content[0].Type != "document" {
		t.Error("type is not document")
	}
	if content[0].Filename != "document.pdf" {
		t.Errorf("expected filename 'document.pdf', got '%s'", content[0].Filename)
	}
	if content[0].MimeType != "application/pdf" {
		t.Error("mime_type is not application/pdf")
	}

	// Test audio message with filename
	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(audioMessageWithFilename))
	if err != nil {
		t.Error(err)
	}
	content = promptMessage.Content.([]PromptMessageContent)
	if content[0].Type != "audio" {
		t.Error("type is not audio")
	}
	if content[0].Filename != "audio_sample.mp3" {
		t.Errorf("expected filename 'audio_sample.mp3', got '%s'", content[0].Filename)
	}

	// Test video message with filename
	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(videoMessageWithFilename))
	if err != nil {
		t.Error(err)
	}
	content = promptMessage.Content.([]PromptMessageContent)
	if content[0].Type != "video" {
		t.Error("type is not video")
	}
	if content[0].Filename != "video_sample.mp4" {
		t.Errorf("expected filename 'video_sample.mp4', got '%s'", content[0].Filename)
	}

	// Test mixed content with filename
	promptMessage, err = parser.UnmarshalJsonBytes[PromptMessage]([]byte(mixedContentWithFilename))
	if err != nil {
		t.Error(err)
	}
	content = promptMessage.Content.([]PromptMessageContent)
	if len(content) != 2 {
		t.Errorf("expected 2 content items, got %d", len(content))
	}
	if content[0].Type != "text" {
		t.Error("first content type is not text")
	}
	if content[1].Type != "image" {
		t.Error("second content type is not image")
	}
	if content[1].Filename != "screenshot.png" {
		t.Errorf("expected filename 'screenshot.png', got '%s'", content[1].Filename)
	}
}

func TestPromptMessageContentWithoutFilename(t *testing.T) {
	const (
		imageWithoutFilename = `
		{
			"role": "user",
			"content": [
				{
					"type": "image",
					"data": "base64encodeddata"
				}
			]
		}`
	)

	// Test that messages without filename still work (backward compatibility)
	promptMessage, err := parser.UnmarshalJsonBytes[PromptMessage]([]byte(imageWithoutFilename))
	if err != nil {
		t.Error(err)
	}
	content := promptMessage.Content.([]PromptMessageContent)
	if content[0].Type != "image" {
		t.Error("type is not image")
	}
	// Filename should be empty string when not provided
	if content[0].Filename != "" {
		t.Errorf("expected empty filename, got '%s'", content[0].Filename)
	}
}
