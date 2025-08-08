package requests

type RequestOAuthGetAuthorizationURL struct {
	Provider          string         `json:"provider" validate:"required"`
	RedirectURI       string         `json:"redirect_uri" validate:"required"`
	SystemCredentials map[string]any `json:"system_credentials" validate:"omitempty"`
}

type RequestOAuthGetCredentials struct {
	Provider          string         `json:"provider" validate:"required"`
	RedirectURI       string         `json:"redirect_uri" validate:"required"`
	SystemCredentials map[string]any `json:"system_credentials" validate:"omitempty"`
	RawHttpRequest    string         `json:"raw_http_request" validate:"required"` // hex encoded raw http request from the oauth provider
}

type RequestOAuthRefreshCredentials struct {
	Provider          string         `json:"provider" validate:"required"`
	RedirectURI       string         `json:"redirect_uri" validate:"required"`
	SystemCredentials map[string]any `json:"system_credentials" validate:"omitempty"`
	Credentials       map[string]any `json:"credentials" validate:"required"`
}
