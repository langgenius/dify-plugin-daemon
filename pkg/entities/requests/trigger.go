package requests

// Request types - matching Python SDK protocol exactly
type TriggerInvokeEventRequest struct {
	Provider       string         `json:"provider" validate:"required"`
	Event          string         `json:"event" validate:"required"`
	RawHTTPRequest string         `json:"raw_http_request" validate:"required"`
	Parameters     map[string]any `json:"parameters" validate:"omitempty"`
	Subscription   map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	Payload        map[string]any `json:"payload" validate:"omitempty"`     // Payload object from `dispatch_event` serialized as dict
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
	Credentials
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
type TriggerInvokeEventResponse struct {
	Variables map[string]any `json:"variables"`
}

type TriggerValidateProviderCredentialsResponse struct {
	Result bool `json:"result"`
}

type TriggerDispatchEventResponse struct {
	UserID   string         `json:"user_id" validate:"omitempty"`
	Events   []string       `json:"events"`
	Payload  map[string]any `json:"payload,omitempty" validate:"omitempty"`
	Response string         `json:"response"`
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
