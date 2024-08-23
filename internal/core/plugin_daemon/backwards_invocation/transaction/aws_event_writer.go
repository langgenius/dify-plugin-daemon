package transaction

import (
	"io"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

// AWSTransactionWriter is a writer that implements the backwards_invocation.BackwardsInvocationWriter interface
// it is used to write data to the plugin runtime
type AWSTransactionWriter struct {
	session     *session_manager.Session
	writeCloser io.WriteCloser

	backwards_invocation.BackwardsInvocationWriter
}

// NewAWSTransactionWriter creates a new transaction writer
func NewAWSTransactionWriter(session *session_manager.Session, writeCloser io.WriteCloser) *AWSTransactionWriter {
	return &AWSTransactionWriter{
		session:     session,
		writeCloser: writeCloser,
	}
}

// Write writes the event and data to the session
// WARNING: write
func (w *AWSTransactionWriter) Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) error {
	_, err := w.writeCloser.Write(w.session.Message(event, data))
	return err
}

func (w *AWSTransactionWriter) Done() {
	w.writeCloser.Close()
}
