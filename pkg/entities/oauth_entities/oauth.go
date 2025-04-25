package oauth_entities

type OAuthGetAuthorizationURLResult struct {
	AuthorizationURL string `json:"authorization_url"`
}

type OAuthGetCredentialsResult struct {
	Credentials map[string]any `json:"credentials"`
}
