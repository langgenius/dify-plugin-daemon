package backwards_invocation

type RequestEvent string

const (
	REQUEST_EVENT_RESPONSE RequestEvent = "backward_invocation_response"
	REQUEST_EVENT_ERROR    RequestEvent = "backward_invocation_error"
	REQUEST_EVENT_END      RequestEvent = "backward_invocation_end"
)

type BaseRequestEvent struct {
	BackwardsRequestId string         `json:"backwards_request_id"`
	Event              RequestEvent   `json:"event"`
	Message            string         `json:"message"`
	Data               map[string]any `json:"data"`
}

func NewResponseEvent(request_id string, message string, data map[string]any) *BaseRequestEvent {
	return &BaseRequestEvent{
		BackwardsRequestId: request_id,
		Event:              REQUEST_EVENT_RESPONSE,
		Message:            message,
		Data:               data,
	}
}

func NewErrorEvent(request_id string, message string) *BaseRequestEvent {
	return &BaseRequestEvent{
		BackwardsRequestId: request_id,
		Event:              REQUEST_EVENT_ERROR,
		Message:            message,
		Data:               nil,
	}
}

func NewEndEvent(request_id string) *BaseRequestEvent {
	return &BaseRequestEvent{
		BackwardsRequestId: request_id,
		Event:              REQUEST_EVENT_END,
		Message:            "",
		Data:               nil,
	}
}
