package plugin_entities

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type PluginUniversalEvent struct {
	SessionId string          `json:"session_id"`
	Event     PluginEventType `json:"event"`
	Data      json.RawMessage `json:"data"`
}

// ParsePluginUniversalEvent parses bytes into struct contains basic info of a message
// it's the outermost layer of the protocol
// error_handler will be called when data is not standard or itself it's an error message
func ParsePluginUniversalEvent(
	data []byte,
	session_handler func(sessionId string, data []byte),
	heartbeat_handler func(),
	error_handler func(err string),
	info_handler func(message string),
) {
	// handle event
	event, err := parser.UnmarshalJsonBytes[PluginUniversalEvent](data)
	if err != nil {
		if len(data) > 1024 {
			error_handler(err.Error() + " original response: " + string(data[:1024]) + "...")
		} else {
			error_handler(err.Error() + " original response: " + string(data))
		}
		return
	}

	sessionId := event.SessionId

	switch event.Event {
	case PLUGIN_EVENT_LOG:
		if event.Event == PLUGIN_EVENT_LOG {
			logEvent, err := parser.UnmarshalJsonBytes[PluginLogEvent](
				event.Data,
			)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				return
			}

			info_handler(logEvent.Message)
		}
	case PLUGIN_EVENT_SESSION:
		session_handler(sessionId, event.Data)
	case PLUGIN_EVENT_ERROR:
		error_handler(string(event.Data))
	case PLUGIN_EVENT_HEARTBEAT:
		heartbeat_handler()
	}
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
	Type SESSION_MESSAGE_TYPE `json:"type" validate:"required"`
	Data json.RawMessage      `json:"data" validate:"required"`
}

type SESSION_MESSAGE_TYPE string

const (
	SESSION_MESSAGE_TYPE_STREAM SESSION_MESSAGE_TYPE = "stream"
	SESSION_MESSAGE_TYPE_END    SESSION_MESSAGE_TYPE = "end"
	SESSION_MESSAGE_TYPE_ERROR  SESSION_MESSAGE_TYPE = "error"
	SESSION_MESSAGE_TYPE_INVOKE SESSION_MESSAGE_TYPE = "invoke"
)

type ErrorResponse struct {
	Message   string         `json:"message"`
	ErrorType string         `json:"error_type"`
	Args      map[string]any `json:"args" validate:"omitempty,max=10"` // max 10 args
}

func (e *ErrorResponse) Error() string {
	return parser.MarshalJson(map[string]any{
		"message":    e.Message,
		"error_type": e.ErrorType,
		"args":       e.Args,
	})
}
