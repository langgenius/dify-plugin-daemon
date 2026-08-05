package model_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func TestTTSResultJSONCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		mimeType string
		expected string
	}{
		{"current", `{"result":"52494646","mime_type":"audio/wav"}`, "audio/wav", `{"result":"52494646","mime_type":"audio/wav"}`},
		{"legacy", `{"result":"52494646"}`, "", `{"result":"52494646"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.UnmarshalJsonBytes[TTSResult]([]byte(tt.payload))
			if err != nil {
				t.Fatalf("unmarshal TTS result: %v", err)
			}
			if result.MimeType != tt.mimeType {
				t.Fatalf("mime type = %q, want %q", result.MimeType, tt.mimeType)
			}
			if got := parser.MarshalJson(result); got != tt.expected {
				t.Fatalf("marshaled result = %s, want %s", got, tt.expected)
			}
		})
	}
}
