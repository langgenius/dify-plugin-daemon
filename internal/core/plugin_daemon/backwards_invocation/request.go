package backwards_invocation

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

type BackwardsInvocationType = dify_invocation.InvokeType

type BackwardsInvocation struct {
	typ              BackwardsInvocationType
	id               string
	detailed_request map[string]any
	session          *session_manager.Session
}

func NewBackwardsInvocation(
	typ BackwardsInvocationType,
	id string, session *session_manager.Session, detailed_request map[string]any,
) *BackwardsInvocation {
	return &BackwardsInvocation{
		typ:              typ,
		id:               id,
		detailed_request: detailed_request,
		session:          session,
	}
}

func (bi *BackwardsInvocation) GetID() string {
	return bi.id
}

func (bi *BackwardsInvocation) WriteError(err error) {
	bi.session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE,
		NewErrorEvent(bi.id, err.Error()),
	)
}

func (bi *BackwardsInvocation) WriteResponse(message string, data any) {
	bi.session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE,
		NewResponseEvent(bi.id, message, data),
	)
}

func (bi *BackwardsInvocation) EndResponse() {
	bi.session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE,
		NewEndEvent(bi.id),
	)
}

func (bi *BackwardsInvocation) Type() BackwardsInvocationType {
	return bi.typ
}

func (bi *BackwardsInvocation) RequestData() map[string]any {
	return bi.detailed_request
}

func (bi *BackwardsInvocation) TenantID() (string, error) {
	if bi.session == nil {
		return "", fmt.Errorf("session is nil")
	}
	return bi.session.TenantID(), nil
}

func (bi *BackwardsInvocation) UserID() (string, error) {
	if bi.session == nil {
		return "", fmt.Errorf("session is nil")
	}
	return bi.session.UserID(), nil
}
