package requests

// Request types - matching Python SDK protocol exactly
type TriggerInvokeRequest struct {
	Provider       string         `json:"provider" validate:"required"`
	Trigger        string         `json:"trigger" validate:"required"`
	RawHTTPRequest string         `json:"raw_http_request" validate:"required"`
	Parameters     map[string]any `json:"parameters" validate:"omitempty"`
	Credentials
}

type TriggerValidateProviderCredentialsRequest struct {
	Provider string `json:"provider" validate:"required"`
	Credentials
}

type TriggerDispatchEventRequest struct {
	Provider       string         `json:"provider" validate:"required"`
	Subscription   map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	RawHTTPRequest string         `json:"raw_http_request" validate:"required"`
}

type TriggerSubscribeRequest struct {
	Provider   string         `json:"provider" validate:"required"`
	Endpoint   string         `json:"endpoint" validate:"required"`
	Parameters map[string]any `json:"parameters" validate:"omitempty"`
	Credentials
}

type TriggerUnsubscribeRequest struct {
	Provider     string         `json:"provider" validate:"required"`
	Subscription map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	Credentials
}

type TriggerRefreshRequest struct {
	Provider     string         `json:"provider" validate:"required"`
	Subscription map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	Credentials
}

// Response types - matching Python SDK protocol exactly
type TriggerInvokeResponse struct {
	Event map[string]any `json:"event"`
}

type TriggerValidateProviderCredentialsResponse struct {
	Result bool `json:"result"`
}

type TriggerDispatchEventResponse struct {
	Triggers        []string `json:"triggers"`
	RawHTTPResponse string   `json:"raw_http_response"`
}

type TriggerSubscribeResponse struct {
	Subscription map[string]any `json:"subscription"`
}

type TriggerUnsubscribeResponse struct {
	Subscription map[string]any `json:"subscription"`
}

type TriggerRefreshResponse struct {
	Subscription map[string]any `json:"subscription"`
}
