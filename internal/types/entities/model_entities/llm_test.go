package model_entities

import (
	"encoding/json"
	"testing"
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

	var prompt_message PromptMessage

	err := json.Unmarshal([]byte(system_message), &prompt_message)
	if err != nil {
		t.Error(err)
	}
	if prompt_message.Role != "system" {
		t.Error("role is not system")
	}

	err = json.Unmarshal([]byte(user_message), &prompt_message)
	if err != nil {
		t.Error(err)
	}
	if prompt_message.Role != "user" {
		t.Error("role is not user")
	}

	err = json.Unmarshal([]byte(assistant_message), &prompt_message)
	if err != nil {
		t.Error(err)
	}
	if prompt_message.Role != "assistant" {
		t.Error("role is not assistant")
	}

	err = json.Unmarshal([]byte(image_message), &prompt_message)
	if err != nil {
		t.Error(err)
	}
	if prompt_message.Role != "user" {
		t.Error("role is not user")
	}
	if prompt_message.Content.([]PromptMessageContent)[0].Type != "image" {
		t.Error("type is not image")
	}

	err = json.Unmarshal([]byte(tool_message), &prompt_message)
	if err != nil {
		t.Error(err)
	}
	if prompt_message.Role != "tool" {
		t.Error("role is not tool")
	}
	if prompt_message.ToolCallId != "123" {
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

	var prompt_message PromptMessage

	err := json.Unmarshal([]byte(wrong_role), &prompt_message)
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

	var prompt_message PromptMessage

	err := json.Unmarshal([]byte(wrong_content), &prompt_message)
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

	var prompt_message PromptMessage

	err := json.Unmarshal([]byte(wrong_content_array), &prompt_message)
	if err == nil {
		t.Error("error is nil")
	}
}

func TestWrongContentArray2(t *testing.T) {
	const (
		wrong_content_array2 = `
		{
			"role": "user",
			"content": [
				{
					"type": "image"
				}
			]
		}
		`
	)

	var prompt_message PromptMessage

	err := json.Unmarshal([]byte(wrong_content_array2), &prompt_message)
	if err == nil {
		t.Error("error is nil")
	}
}

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

	var prompt_message PromptMessage

	err := json.Unmarshal([]byte(wrong_content_array3), &prompt_message)
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

	var c LLMResultChunk

	err := json.Unmarshal([]byte(llm_result_chunk), &c)
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

	var u LLMUsage

	err := json.Unmarshal([]byte(llm_usage), &u)
	if err != nil {
		t.Error(err)
	}
}
