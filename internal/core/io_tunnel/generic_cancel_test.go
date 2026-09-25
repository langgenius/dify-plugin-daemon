package io_tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

var errCancelAfterTeardown = errors.New("cancel reached the runtime after the listener was torn down")

type fullDuplexRuntime struct {
	mu          sync.Mutex
	runtimeType plugin_entities.PluginRuntimeType
	listener    *entities.Broadcast[plugin_entities.SessionMessage]
	listening   bool
	events      []string
	writeErr    error
}

func (r *fullDuplexRuntime) Type() plugin_entities.PluginRuntimeType {
	if r.runtimeType != "" {
		return r.runtimeType
	}
	return plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
}

func (r *fullDuplexRuntime) Configuration() *plugin_entities.PluginDeclaration {
	return &plugin_entities.PluginDeclaration{}
}

func (r *fullDuplexRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	return testPluginUniqueIdentifier, nil
}

func (r *fullDuplexRuntime) HashedIdentity() (string, error) {
	return plugin_entities.HashedIdentity(testPluginUniqueIdentifier.String()), nil
}

func (r *fullDuplexRuntime) Checksum() (string, error) {
	return testPluginUniqueIdentifier.Checksum(), nil
}

func (r *fullDuplexRuntime) Listen(string) (*entities.Broadcast[plugin_entities.SessionMessage], error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.listener = entities.NewCallbackHandler[plugin_entities.SessionMessage]()
	r.listening = true
	r.listener.OnClose(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.listening = false
	})
	return r.listener, nil
}

func (r *fullDuplexRuntime) Write(_ string, _ access_types.PluginAccessAction, data []byte) error {
	var message struct {
		Event string `json:"event"`
	}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if message.Event == string(session_manager.PLUGIN_IN_STREAM_EVENT_CANCEL) && !r.listening {
		return errCancelAfterTeardown
	}
	r.events = append(r.events, message.Event)
	return r.writeErr
}

func (r *fullDuplexRuntime) send(message plugin_entities.SessionMessage) {
	r.mu.Lock()
	listener := r.listener
	r.mu.Unlock()
	listener.Send(message)
}

func (r *fullDuplexRuntime) recordedEvents() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.events...)
}

func startInvocation(
	t *testing.T,
	runtime plugin_entities.PluginRuntimeSessionIOInterface,
) *session_manager.Session {
	t.Helper()
	routine.InitPool(4)

	session := session_manager.NewSession(session_manager.NewSessionPayload{
		TenantID:               "tenant-1",
		UserID:                 "user-1",
		PluginUniqueIdentifier: testPluginUniqueIdentifier,
		InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_MODEL,
		Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
		RequestContext:         context.Background(),
		IgnoreCache:            true,
	})
	session.BindRuntime(runtime)
	t.Cleanup(func() {
		session.Close(session_manager.CloseSessionPayload{IgnoreCache: true})
	})

	return session
}

func TestAbandonedInvocationIsCancelledOnThePlugin(t *testing.T) {
	runtime := &fullDuplexRuntime{}
	session := startInvocation(t, runtime)

	response, err := GenericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		session, &requests.RequestInvokeLLM{}, 8,
	)
	require.NoError(t, err)
	require.Equal(t, []string{"request"}, runtime.recordedEvents())

	response.Close()

	require.Equal(t, []string{"request", "cancel"}, runtime.recordedEvents())
}

func TestCompletedInvocationIsNotCancelled(t *testing.T) {
	runtime := &fullDuplexRuntime{}
	session := startInvocation(t, runtime)

	response, err := GenericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		session, &requests.RequestInvokeLLM{}, 8,
	)
	require.NoError(t, err)

	runtime.send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_END})
	require.False(t, response.Next())

	response.Close()

	require.Equal(t, []string{"request"}, runtime.recordedEvents())
}

func TestInvocationThatFailedInThePluginIsNotCancelled(t *testing.T) {
	runtime := &fullDuplexRuntime{}
	session := startInvocation(t, runtime)

	response, err := GenericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		session, &requests.RequestInvokeLLM{}, 8,
	)
	require.NoError(t, err)

	runtime.send(plugin_entities.SessionMessage{
		Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
		Data: []byte(`{"error_type":"InvokeError","message":"nope","args":{}}`),
	})
	require.True(t, response.Next())
	_, readErr := response.Read()
	require.Error(t, readErr)

	response.Close()

	require.Equal(t, []string{"request"}, runtime.recordedEvents())
}

func TestServerlessInvocationIsNotCancelledOverTheRequestChannel(t *testing.T) {
	runtime := &fullDuplexRuntime{runtimeType: plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS}
	session := startInvocation(t, runtime)

	response, err := GenericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		session, &requests.RequestInvokeLLM{}, 8,
	)
	require.NoError(t, err)

	response.Close()

	require.Equal(t, []string{"request"}, runtime.recordedEvents())
}
