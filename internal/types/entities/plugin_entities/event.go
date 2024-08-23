package plugin_entities

import (
	"encoding/json"
	"io"
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
	Type        SESSION_MESSAGE_TYPE `json:"type"`
	Data        json.RawMessage      `json:"data"`
	RuntimeType PluginRuntimeType    `json:"runtime_type"`

	// only used for aws event bridge, not used for stdio and tcp
	SessionWriter io.WriteCloser `json:"-"`
}

type SESSION_MESSAGE_TYPE string

const (
	SESSION_MESSAGE_TYPE_STREAM SESSION_MESSAGE_TYPE = "stream"
	SESSION_MESSAGE_TYPE_END    SESSION_MESSAGE_TYPE = "end"
	SESSION_MESSAGE_TYPE_ERROR  SESSION_MESSAGE_TYPE = "error"
	SESSION_MESSAGE_TYPE_INVOKE SESSION_MESSAGE_TYPE = "invoke"
)

type PluginResponseChunk struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
