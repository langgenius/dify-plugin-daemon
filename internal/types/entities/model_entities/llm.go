package model_entities

type ModelType string

const (
	MODEL_TYPE_LLM            ModelType = "llm"
	MODEL_TYPE_TEXT_EMBEDDING ModelType = "text_embedding"
	MODEL_TYPE_RERANKING      ModelType = "rerank"
)

type LLMModel string

const (
	LLM_MODE_CHAT       LLMModel = "chat"
	LLM_MODE_COMPLETION LLMModel = "completion"
)
