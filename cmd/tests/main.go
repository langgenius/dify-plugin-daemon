package main

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
)

func main() {
	const data = `{"data":{"structured_output":{"name":"yeuoly","age":0},"model":"gpt-4o-mini","prompt_messages":[],"system_fingerprint":"fp_34a54ae93c","delta":{"index":0,"message":{"role":"assistant","content":"","name":null,"tool_calls":[]},"usage":null,"finish_reason":null}},"error":""}`
	type resp struct {
		Data model_entities.LLMResultChunkWithStructuredOutput `json:"data"`
	}
	output, err := parser.UnmarshalJson[resp](data)
	if err != nil {
		panic(err)
	}

	fmt.Println(parser.MarshalJson(output.Data))
}
