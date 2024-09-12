package transaction

import (
	"io"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

type WriteFlushCloser interface {
	io.WriteCloser

	Flush()
}

// AWSTransactionWriter is a writer that implements the backwards_invocation.BackwardsInvocationWriter interface
// it is used to write data to the plugin runtime
type AWSTransactionWriter struct {
	session          *session_manager.Session
	writeFlushCloser WriteFlushCloser

	backwards_invocation.BackwardsInvocationWriter
}

// NewAWSTransactionWriter creates a new transaction writer
func NewAWSTransactionWriter(
	session *session_manager.Session,
	writeFlushCloser WriteFlushCloser,
) *AWSTransactionWriter {
	return &AWSTransactionWriter{
		session:          session,
		writeFlushCloser: writeFlushCloser,
	}
}

// Write writes the event and data to the session
func (w *AWSTransactionWriter) Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) error {
	_, err := w.writeFlushCloser.Write(append(w.session.Message(event, data), '\n', '\n'))
	if err != nil {
		return err
	}
	w.writeFlushCloser.Flush()
	return err
}

func (w *AWSTransactionWriter) Done() {
	w.writeFlushCloser.Close()
}
