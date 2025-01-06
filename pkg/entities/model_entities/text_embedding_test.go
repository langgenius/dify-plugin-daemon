package model_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func TestTextEmbeddingFullFunction(t *testing.T) {
	const (
		text_embedding = `
		{
			"model": "text_embedding",
			"embeddings": [[
				0.1, 0.2, 0.3
			]],
			"usage" : {
				"tokens": 3,
				"total_tokens": 100,
				"unit_price": 0.1,
				"price_unit": 1,
				"total_price": 10,
				"currency": "usd",
				"latency": 0.1
			}
		}`
	)

	_, err := parser.UnmarshalJsonBytes[TextEmbeddingResult]([]byte(text_embedding))
	if err != nil {
		t.Error(err)
	}
}

func TestTextEmbeddingWrongUsage(t *testing.T) {
	const (
		text_embedding = `
		{
			"model": "text_embedding",
			"embeddings": [[
				0.1, 0.2, 0.3
			]],
			"usage" : {
				"tokens": 3,
				"total_tokens": 100,
				"unit_price": 0.1,
				"price_unit": 1,
				"total_price": 10,
				"currency": "usd"
			}
		}`
	)

	_, err := parser.UnmarshalJsonBytes[TextEmbeddingResult]([]byte(text_embedding))
	if err == nil {
		t.Error("should have error")
	}
}

func TestTextEmbeddingWrongEmbeddings(t *testing.T) {
	const (
		text_embedding = `
		{
			"model": "text_embedding",
			"embeddings": [
				0.1, 0.2, 0.3
			],
			"usage" : {
				"tokens": 3,
				"total_tokens": 100,
				"unit_price": 0.1,
				"price_unit": 1,
				"total_price": 10,
				"currency": "usd",
				"latency": 0.1
			}
		}`
	)

	_, err := parser.UnmarshalJsonBytes[TextEmbeddingResult]([]byte(text_embedding))
	if err == nil {
		t.Error("should have error")
	}
}
