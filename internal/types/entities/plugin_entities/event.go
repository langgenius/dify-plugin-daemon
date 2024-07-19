package plugin_entities

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
)

type PluginUniversalEvent struct {
	Event     PluginEventType `json:"event"`
	SessionId string          `json:"session_id"`
	Data      json.RawMessage `json:"data"`
}

type PluginEventType string

const (
	PLUGIN_EVENT_LOG       PluginEventType = "log"
	PLUGIN_EVENT_SESSION   PluginEventType = "session"
	PLUGIN_EVENT_ERROR     PluginEventType = "error"
	PLUGIN_EVENT_HEARTBEAT PluginEventType = "heartbeat"
)

type PluginLogEvent struct {
	Level     string  `json:"level"`
	Message   string  `json:"message"`
	Timestamp float64 `json:"timestamp"`
}

type SessionMessage struct {
	Type SESSION_MESSAGE_TYPE `json:"type"`
	Data json.RawMessage      `json:"data"`
}

type SESSION_MESSAGE_TYPE string

const (
	SESSION_MESSAGE_TYPE_STREAM SESSION_MESSAGE_TYPE = "stream"
	SESSION_MESSAGE_TYPE_END    SESSION_MESSAGE_TYPE = "end"
	SESSION_MESSAGE_TYPE_ERROR  SESSION_MESSAGE_TYPE = "error"
	SESSION_MESSAGE_TYPE_INVOKE SESSION_MESSAGE_TYPE = "invoke"
)

type ToolResponseChunk struct {
	Type    string         `json:"type"`
	Message map[string]any `json:"message"`
}

type PluginResponseChunk struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type InvokeModelResponseChunk = model_entities.LLMResultChunk

type ErrorResponse struct {
	Error string `json:"error"`
}
