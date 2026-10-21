package plugin_entities

import (
	"bytes"
	"encoding/json"
)

type InvokePluginUserIdentity struct {
	TenantId string `json:"tenant_id" validate:"required" uri:"tenant_id"`
	UserId   string `json:"user_id"`
}

type BasePluginIdentifier struct {
	PluginID string `json:"plugin_id"`
}

type InvokePluginRequest[T any] struct {
	InvokePluginUserIdentity
	BasePluginIdentifier

	UniqueIdentifier PluginUniqueIdentifier `json:"unique_identifier"`
	ConversationID   *string                `json:"conversation_id"`
	MessageID        *string                `json:"message_id"`
	AppID            *string                `json:"app_id"`
	EndpointID       *string                `json:"endpoint_id"`
	Context          map[string]any         `json:"context"`

	Data T `json:"data" validate:"required"`
}

func (r *InvokePluginRequest[T]) UnmarshalJSON(data []byte) error {
	type request InvokePluginRequest[T]
	decoder := json.NewDecoder(bytes.NewReader(data))
	// Plugin arguments and saved model state must retain their original JSON numbers.
	decoder.UseNumber()
	return decoder.Decode((*request)(r))
}
