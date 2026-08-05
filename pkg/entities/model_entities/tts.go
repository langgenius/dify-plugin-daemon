package model_entities

type TTSResult struct {
	Result   string `json:"result"` // in hex
	MimeType string `json:"mime_type,omitempty"`
}

type TTSModelVoice struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type GetTTSVoicesResponse struct {
	Voices []TTSModelVoice `json:"voices" validate:"required,dive"`
}
