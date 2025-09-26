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

// ServerlessTransactionWriter is a writer that implements the backwards_invocation.BackwardsInvocationWriter interface
// it is used to write data to the plugin runtime
type ServerlessTransactionWriter struct {
	session          *session_manager.Session
	writeFlushCloser WriteFlushCloser

	backwards_invocation.BackwardsInvocationWriter
}

// NewServerlessTransactionWriter creates a new transaction writer
func NewServerlessTransactionWriter(
	session *session_manager.Session,
	writeFlushCloser WriteFlushCloser,
) *ServerlessTransactionWriter {
	return &ServerlessTransactionWriter{
		session:          session,
		writeFlushCloser: writeFlushCloser,
	}
}

// Write writes the event and data to the session
func (w *ServerlessTransactionWriter) Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) error {
	_, err := w.writeFlushCloser.Write(append(w.session.Message(event, data), '\n', '\n'))
	if err != nil {
		return err
	}
	w.writeFlushCloser.Flush()
	return err
}

func (w *ServerlessTransactionWriter) Done() {
	w.writeFlushCloser.Close()
}
