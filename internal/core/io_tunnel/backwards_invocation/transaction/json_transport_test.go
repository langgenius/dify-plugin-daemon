package transaction

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/stretchr/testify/require"
)

type replayInvocation struct {
	dify_invocation.BackwardsInvocation
	request *dify_invocation.InvokeLLMRequest
}

func (i *replayInvocation) SetContext(context.Context) {}

func (i *replayInvocation) InvokeLLM(request *dify_invocation.InvokeLLMRequest) (*stream.Stream[model_entities.LLMResultChunk], error) {
	i.request = request
	result := stream.NewStream[model_entities.LLMResultChunk](1)
	index := 0
	result.Write(model_entities.LLMResultChunk{
		Model: "model",
		Delta: model_entities.LLMResultChunkDelta{Index: &index, Message: request.PromptMessages[0]},
	})
	result.Close()
	return result, nil
}

type replayResponseWriter struct {
	bytes.Buffer
	done chan struct{}
}

func (w *replayResponseWriter) Flush()       {}
func (w *replayResponseWriter) Close() error { close(w.done); return nil }

func TestServerlessInvocationPreservesRawJSON(t *testing.T) {
	routine.InitPool(4)
	invocation := &replayInvocation{}
	session := session_manager.NewSession(session_manager.NewSessionPayload{
		TenantID: "tenant", UserID: "user", BackwardsInvocation: invocation, IgnoreCache: true,
	})
	t.Cleanup(func() {
		session_manager.DeleteSession(session_manager.DeleteSessionPayload{ID: session.ID, IgnoreCache: true})
	})
	declaration := &plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Resource: plugin_entities.PluginResourceRequirement{
				Permission: &plugin_entities.PluginPermissionRequirement{
					Model: &plugin_entities.PluginPermissionModelRequirement{Enabled: true, LLM: true},
				},
			},
		},
	}
	writer := &replayResponseWriter{done: make(chan struct{})}
	body, err := json.Marshal(plugin_entities.PluginUniversalEvent{
		SessionId: session.ID,
		Event:     plugin_entities.PLUGIN_EVENT_SESSION,
		Data:      json.RawMessage(`{"type":"invoke","data":{"type":"llm","backwards_request_id":"request","request":{"provider":"p","model":"m","mode":"chat","completion_params":{"seed":9007199254740993},"prompt_messages":[{"role":"assistant","content":null,"opaque_body":{"id":9007199254740993}}]}}}`),
	})
	require.NoError(t, err)
	// Follow the serverless handler's event parsing and invoke path with a controlled session.
	plugin_entities.ParsePluginUniversalEvent(body, "", func(sessionID string, data []byte) {
		require.Equal(t, session.ID, sessionID)
		message, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
		require.NoError(t, err)
		require.NoError(t, backwards_invocation.InvokeDify(
			declaration, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, session,
			NewServerlessTransactionWriter(session, writer), message.Data,
		))
	}, nil, func(err string) {
		t.Fatalf("unexpected event error: %s", err)
	}, nil)
	select {
	case <-writer.done:
	case <-time.After(5 * time.Second):
		t.Fatal("serverless writer did not close")
	}
	require.NotNil(t, invocation.request)
	require.Equal(t, json.Number("9007199254740993"), invocation.request.CompletionParams["seed"])
	require.Equal(t, `{"id":9007199254740993}`, string(invocation.request.PromptMessages[0].OpaqueBody))
	require.Contains(t, writer.String(), `"opaque_body":{"id":9007199254740993}`)
	require.Contains(t, writer.String(), `"event":"end"`)
}
