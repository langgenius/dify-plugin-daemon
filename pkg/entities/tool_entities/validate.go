package tool_entities

type ValidateCredentialsResult struct {
	Result                 *bool `json:"result,omitempty" validate:"required_without=CredentialCheckResult,excluded_with=CredentialCheckResult"`
	*CredentialCheckResult `validate:"required_without=Result,excluded_with=Result"`
}

type CredentialCheckResult struct {
	Status     string  `json:"status" validate:"required,oneof=valid invalid temporarily_unavailable unsupported"`
	ReasonCode *string `json:"reason_code"`
	Retryable  bool    `json:"retryable"`
}
