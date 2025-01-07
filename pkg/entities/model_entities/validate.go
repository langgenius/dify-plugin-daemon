package model_entities

type ValidateCredentialsResult struct {
	Result      bool           `json:"result"`
	Credentials map[string]any `json:"credentials"`
}
