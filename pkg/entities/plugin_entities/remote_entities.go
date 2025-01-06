package plugin_entities

import "encoding/json"

type RemoteAssetPayload struct {
	Filename string `json:"filename" validate:"required"`
	Data     string `json:"data" validate:"required"`
}

type RemotePluginRegisterEventType string

const (
	REGISTER_EVENT_TYPE_HAND_SHAKE                 RemotePluginRegisterEventType = "handshake"
	REGISTER_EVENT_TYPE_ASSET_CHUNK                RemotePluginRegisterEventType = "asset_chunk"
	REGISTER_EVENT_TYPE_MANIFEST_DECLARATION       RemotePluginRegisterEventType = "manifest_declaration"
	REGISTER_EVENT_TYPE_TOOL_DECLARATION           RemotePluginRegisterEventType = "tool_declaration"
	REGISTER_EVENT_TYPE_MODEL_DECLARATION          RemotePluginRegisterEventType = "model_declaration"
	REGISTER_EVENT_TYPE_ENDPOINT_DECLARATION       RemotePluginRegisterEventType = "endpoint_declaration"
	REGISTER_EVENT_TYPE_AGENT_STRATEGY_DECLARATION RemotePluginRegisterEventType = "agent_strategy_declaration"
	REGISTER_EVENT_TYPE_END                        RemotePluginRegisterEventType = "end"
)

type RemotePluginRegisterAssetChunk struct {
	Filename string `json:"filename" validate:"required"`
	Data     string `json:"data" validate:"required"`
	End      bool   `json:"end"` // if true, it's the last chunk of the file
}

type RemotePluginRegisterHandshake struct {
	Key string `json:"key" validate:"required"`
}

type RemotePluginRegisterPayload struct {
	Type RemotePluginRegisterEventType `json:"type" validate:"required"`
	Data json.RawMessage               `json:"data" validate:"required"`
}
