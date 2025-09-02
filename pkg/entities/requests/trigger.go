package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type TriggerActions string

const (
	TriggerActionsInvokeTrigger               TriggerActions = "invoke_trigger"
	TriggerActionsValidateProviderCredentials TriggerActions = "validate_provider_credentials"
	TriggerActionsDispatchTriggerEvent        TriggerActions = "dispatch_trigger_event"
	TriggerActionsSubscribeTrigger            TriggerActions = "subscribe_trigger"
	TriggerActionsUnsubscribeTrigger          TriggerActions = "unsubscribe_trigger"
	TriggerActionsRefreshTrigger              TriggerActions = "refresh_trigger"
)

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("trigger_action", func(fl validator.FieldLevel) bool {
		switch fl.Field().String() {
		case string(TriggerActionsInvokeTrigger),
			string(TriggerActionsValidateProviderCredentials),
			string(TriggerActionsDispatchTriggerEvent),
			string(TriggerActionsSubscribeTrigger),
			string(TriggerActionsUnsubscribeTrigger),
			string(TriggerActionsRefreshTrigger):
			return true
		}
		return false
	})
}

// Request types - matching Python SDK protocol exactly
type TriggerInvokeRequest struct {
	Action         TriggerActions `json:"action" validate:"required,trigger_action"`
	Provider       string         `json:"provider" validate:"required"`
	Trigger        string         `json:"trigger" validate:"required"`
	RawHTTPRequest string         `json:"raw_http_request" validate:"required"`
	Parameters     map[string]any `json:"parameters" validate:"omitempty"`
	Credentials
}

type TriggerValidateProviderCredentialsRequest struct {
	Action   TriggerActions `json:"action" validate:"required,trigger_action"`
	Provider string         `json:"provider" validate:"required"`
	Credentials
}

type TriggerDispatchEventRequest struct {
	Action         TriggerActions `json:"action" validate:"required,trigger_action"`
	Provider       string         `json:"provider" validate:"required"`
	Subscription   map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	RawHTTPRequest string         `json:"raw_http_request" validate:"required"`
}

type TriggerSubscribeRequest struct {
	Action     TriggerActions `json:"action" validate:"required,trigger_action"`
	Provider   string         `json:"provider" validate:"required"`
	Endpoint   string         `json:"endpoint" validate:"required"`
	Parameters map[string]any `json:"parameters" validate:"omitempty"`
	Credentials
}

type TriggerUnsubscribeRequest struct {
	Action       TriggerActions `json:"action" validate:"required,trigger_action"`
	Provider     string         `json:"provider" validate:"required"`
	Subscription map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	Credentials
}

type TriggerRefreshRequest struct {
	Action       TriggerActions `json:"action" validate:"required,trigger_action"`
	Provider     string         `json:"provider" validate:"required"`
	Subscription map[string]any `json:"subscription" validate:"required"` // Subscription object serialized as dict
	Credentials
}

// Response types - matching Python SDK protocol exactly
type TriggerInvokeResponse struct {
	Event map[string]any `json:"event"`
}

type TriggerValidateProviderCredentialsResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
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
