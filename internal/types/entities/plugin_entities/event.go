package plugin_entities

import "encoding/json"

type PluginUniversalEvent struct {
	Event     string          `json:"event"`
	SessionId string          `json:"session_id"`
	Data      json.RawMessage `json:"data"`
}

const (
	PLUGIN_EVENT_LOG      = "log"
	PLUGIN_EVENT_RESPONSE = "response"
	PLUGIN_EVENT_ERROR    = "error"
	PLUGIN_EVENT_INVOKE   = "invoke"
)

type PluginLogEvent struct {
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	Timestamp float64 `json:"timestamp"`
}

type StreamMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

const (
	STREAM_MESSAGE_TYPE_STREAM = "stream"
	STREAM_MESSAGE_TYPE_END    = "end"
)

type InvokeToolResponseChunk struct {
	Type    string          `json:"type" binding:"required"`
	Message json.RawMessage `json:"message" binding:"required"`
}

type InvokeModelResponseChunk struct {
}
