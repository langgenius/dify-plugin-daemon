package oauth_entities

type OAuthGetAuthorizationURLResult struct {
	AuthorizationURL string `json:"authorization_url"`
}

type OAuthGetCredentialsResult struct {
	Metadata    map[string]any `json:"metadata,omitempty"`
	Credentials map[string]any `json:"credentials"`
	ExpiresAt   uint32         `json:"expires_at"`
}
