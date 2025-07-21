package oauth_entities

type OAuthGetAuthorizationURLResult struct {
	AuthorizationURL string `json:"authorization_url"`
}

type OAuthGetCredentialsResult struct {
	Credentials map[string]any `json:"credentials"`
	ExpiresAt   int64          `json:"expires_at"`
}

type OAuthRefreshCredentialsResult struct {
	Credentials map[string]any `json:"credentials"`
	ExpiresAt   int64          `json:"expires_at"`
}
