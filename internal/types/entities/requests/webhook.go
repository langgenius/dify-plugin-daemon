package requests

type RequestInvokeWebhook struct {
	RawHttpRequest string `json:"raw_http_request" validate:"required"`
}
