package transaction

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

// AWSTransactionWriter is a writer that implements the backwards_invocation.BackwardsInvocationWriter interface
// it is used to write data to the plugin runtime
type AWSTransactionWriter struct {
	event_id string

	backwards_invocation.BackwardsInvocationWriter
}

// NewAWSTransactionWriter creates a new transaction writer
func NewAWSTransactionWriter(event_id string) *AWSTransactionWriter {
	return &AWSTransactionWriter{
		event_id: event_id,
	}
}

func (w *AWSTransactionWriter) Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) {

}

func (w *AWSTransactionWriter) Done() {

}
