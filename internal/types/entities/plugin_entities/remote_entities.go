package plugin_entities

type RemoteAssetPayload struct {
	Filename string `json:"filename" validate:"required"`
	Data     string `json:"data" validate:"required"`
}
