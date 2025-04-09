package transaction

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

// FullDuplexTransactionWriter is a writer that implements the backwards_invocation.BackwardsInvocationWriter interface
// write data into session
type FullDuplexTransactionWriter struct {
	session *session_manager.Session

	backwards_invocation.BackwardsInvocationWriter
}

func NewFullDuplexEventWriter(session *session_manager.Session) *FullDuplexTransactionWriter {
	return &FullDuplexTransactionWriter{
		session: session,
	}
}

func (w *FullDuplexTransactionWriter) Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) error {
	return w.session.Write(event, "", data)
}

func (w *FullDuplexTransactionWriter) Done() {
}
