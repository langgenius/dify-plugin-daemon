package requests

type RequestDynamicParameterSelect struct {
	Credentials    map[string]any `json:"credentials" validate:"required"`
	Provider       string         `json:"provider" validate:"required"`
	ProviderAction string         `json:"provider_action" validate:"required"`
	Parameter      string         `json:"parameter" validate:"required"`
}
