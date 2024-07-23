package backwards_invocation

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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
	bi.session.Write(parser.MarshalJsonBytes(NewErrorEvent(bi.id, err.Error())))
}

func (bi *BackwardsInvocation) Write(message string, data map[string]any) {
	bi.session.Write(parser.MarshalJsonBytes(NewResponseEvent(bi.id, message, data)))
}

func (bi *BackwardsInvocation) End() {
	bi.session.Write(parser.MarshalJsonBytes(NewEndEvent(bi.id)))
}

func (bi *BackwardsInvocation) Type() BackwardsInvocationType {
	return bi.typ
}

func (bi *BackwardsInvocation) RequestData() map[string]any {
	return bi.detailed_request
}