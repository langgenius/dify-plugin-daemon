package model_entities

type RerankDocument struct {
	Index *int     `json:"index" validate:"required"`
	Text  *string  `json:"text" validate:"required"`
	Score *float64 `json:"score" validate:"required"`
}

type RerankResult struct {
	Model string           `json:"model" validate:"required"`
	Docs  []RerankDocument `json:"docs" validate:"required,dive"`
}
