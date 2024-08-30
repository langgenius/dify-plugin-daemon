package requests

type RequestInvokeEndpoint struct {
	RawHttpRequest string         `json:"raw_http_request" validate:"required"`
	Settings       map[string]any `json:"settings" validate:"omitempty"`
}
