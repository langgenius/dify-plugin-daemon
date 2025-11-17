package backwards_invocation

type RequestEvent string

const (
	REQUEST_EVENT_RESPONSE RequestEvent = "response"
	REQUEST_EVENT_ERROR    RequestEvent = "error"
	REQUEST_EVENT_END      RequestEvent = "end"
)

type BackwardsInvocationResponseEvent struct {
	BackwardsRequestId string       `json:"backwards_request_id"`
	Event              RequestEvent `json:"event"`
	Message            string       `json:"message"`
	Data               any          `json:"data"`
}

func NewResponseEvent(request_id string, message string, data any) *BackwardsInvocationResponseEvent {
	return &BackwardsInvocationResponseEvent{
		BackwardsRequestId: request_id,
		Event:              REQUEST_EVENT_RESPONSE,
		Message:            message,
		Data:               data,
	}
}

func NewErrorEvent(request_id string, message string) *BackwardsInvocationResponseEvent {
	return &BackwardsInvocationResponseEvent{
		BackwardsRequestId: request_id,
		Event:              REQUEST_EVENT_ERROR,
		Message:            message,
		Data:               nil,
	}
}

func NewEndEvent(request_id string) *BackwardsInvocationResponseEvent {
	return &BackwardsInvocationResponseEvent{
		BackwardsRequestId: request_id,
		Event:              REQUEST_EVENT_END,
		Message:            "",
		Data:               nil,
	}
}
