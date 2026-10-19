package oauth_entities

type OAuthGetAuthorizationURLResult struct {
	AuthorizationURL string `json:"authorization_url"`
}

type OAuthGetCredentialsResult struct {
	Metadata    map[string]any `json:"metadata,omitempty"`
	Credentials map[string]any `json:"credentials"`
	ExpiresAt   int64          `json:"expires_at"`
}

type OAuthRefreshCredentialsResult struct {
	Credentials map[string]any `json:"credentials"`
	ExpiresAt   int64          `json:"expires_at"`
}

type OAuthRevokeCredentialsResult struct {
	Status     string  `json:"status" validate:"required,oneof=revoked unsupported temporarily_unavailable failed"`
	ReasonCode *string `json:"reason_code"`
	Retryable  bool    `json:"retryable"`
}
