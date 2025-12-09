package backwards_invocation

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

type BackwardsInvocationType = dify_invocation.InvokeType

type BackwardsInvocationWriter interface {
	Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) error
	Done()
}

// BackwardsInvocation is a struct that represents a backwards invocation
// For different plugin runtime type, stream handler is different
//  1. Local and Remote: they are both full duplex, multiplexing could be implemented by different session
//     different session share the same physical channel.
//  2. Serverless: it is half duplex, one request could have multiple channels, we need to combine them into one stream
//
// That's why it has a writer, for different transaction, the writer is unique
type BackwardsInvocation struct {
	typ             BackwardsInvocationType
	id              string
	detailedRequest map[string]any
	session         *session_manager.Session

	// writer is the writer that writes the data to the session
	// NOTE: write operation will not raise errors
	writer BackwardsInvocationWriter

	// backwardsInvocation is the backwards invocation that is used to invoke dify
	backwardsInvocation dify_invocation.BackwardsInvocation
}

func NewBackwardsInvocation(
	typ BackwardsInvocationType,
	id string,
	session *session_manager.Session,
	writer BackwardsInvocationWriter,
	detailedRequest map[string]any,
) *BackwardsInvocation {
	return &BackwardsInvocation{
		typ:                 typ,
		id:                  id,
		detailedRequest:     detailedRequest,
		session:             session,
		writer:              writer,
		backwardsInvocation: session.BackwardsInvocation(),
	}
}

func (bi *BackwardsInvocation) GetID() string {
	return bi.id
}

func (bi *BackwardsInvocation) WriteError(err error) {
	bi.writer.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE,
		NewErrorEvent(bi.id, err.Error()),
	)
}

func (bi *BackwardsInvocation) WriteResponse(message string, data any) {
	bi.writer.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE,
		NewResponseEvent(bi.id, message, data),
	)
}

func (bi *BackwardsInvocation) EndResponse() {
	bi.writer.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE,
		NewEndEvent(bi.id),
	)
	bi.writer.Done()
}

func (bi *BackwardsInvocation) Type() BackwardsInvocationType {
	return bi.typ
}

func (bi *BackwardsInvocation) RequestData() map[string]any {
	return bi.detailedRequest
}

func (bi *BackwardsInvocation) TenantID() (string, error) {
	if bi.session == nil {
		return "", fmt.Errorf("session is nil")
	}
	return bi.session.TenantID, nil
}

func (bi *BackwardsInvocation) UserID() (string, error) {
	if bi.session == nil {
		return "", fmt.Errorf("session is nil")
	}
	return bi.session.UserID, nil
}
