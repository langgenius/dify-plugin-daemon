package model_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func TestRerankFullFunction(t *testing.T) {
	const (
		rerank = `
		{
			"model": "rerank",
			"docs": [
				{
					"index": 1,
					"text": "text",
					"score": 0.1
				}
			]
		}`
	)

	_, err := parser.UnmarshalJsonBytes[RerankResult]([]byte(rerank))
	if err != nil {
		t.Error(err)
	}
}

func TestRerankWrongDocs(t *testing.T) {
	const (
		rerank = `
		{
			"model": "rerank",
			"docs": [
				{
					"index": 1,
					"text": "text"
				}
			]
		}`
	)

	_, err := parser.UnmarshalJsonBytes[RerankResult]([]byte(rerank))
	if err == nil {
		t.Error("should have error")
	}
}

func TestRerankWrongDocIndex(t *testing.T) {
	const (
		rerank = `
		{
			"model": "rerank",
			"docs": [
				{
					"text": "text",
					"score": 0.1
				}
			]
		}`
	)

	_, err := parser.UnmarshalJsonBytes[RerankResult]([]byte(rerank))
	if err == nil {
		t.Error("should have error")
	}
}
