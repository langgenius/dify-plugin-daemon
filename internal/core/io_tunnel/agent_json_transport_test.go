package io_tunnel

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

func TestAgentPreservesHistoryAndBlobResponses(t *testing.T) {
	routine.InitPool(4)
	for _, runtimeType := range []plugin_entities.PluginRuntimeType{
		plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
	} {
		t.Run(string(runtimeType), func(t *testing.T) {
			listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
			runtime := &fakePluginRuntime{
				listener: listener, uniqueIdentifier: testPluginUniqueIdentifier, runtimeType: runtimeType,
			}
			runtime.writeFn = func(_ string, _ access_types.PluginAccessAction, _ []byte) error {
				for _, payload := range []string{
					`{"type":"json","message":{"json_object":{"model_history":{"n":9007199254740993,"state":[false,0,{},[]]}}}}`,
					`{"type":"blob_chunk","message":{"id":"file","total_length":2,"blob":"aGk=","end":false}}`,
					`{"type":"blob_chunk","message":{"id":"file","total_length":2,"blob":"","end":true}}`,
				} {
					listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM, Data: []byte(payload)})
				}
				listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_END})
				return nil
			}
			session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_AGENT_STRATEGY,
				access_types.PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY)
			response, err := InvokeAgentStrategy(session, &requests.RequestInvokeAgentStrategy{})
			require.NoError(t, err)
			defer response.Close()
			require.True(t, response.Next())
			history, err := response.Read()
			require.NoError(t, err)
			encoded, err := json.Marshal(history)
			require.NoError(t, err)
			require.Contains(t, string(encoded), `"n":9007199254740993`)
			require.Contains(t, string(encoded), `"state":[false,0,{},[]]`)
			require.True(t, response.Next())
			blob, err := response.Read()
			require.NoError(t, err)
			require.Equal(t, tool_entities.ToolResponseChunkTypeBlob, blob.Type)
			require.Equal(t, []byte("hi"), blob.Message["blob"])
			require.False(t, response.Next())
		})
	}
}

func TestAgentRejectsInvalidBlobLengths(t *testing.T) {
	routine.InitPool(4)
	for _, length := range []string{"-1", "15728641", "1.5", "9223372036854775808"} {
		t.Run(length, func(t *testing.T) {
			listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
			runtime := &fakePluginRuntime{listener: listener, uniqueIdentifier: testPluginUniqueIdentifier}
			runtime.writeFn = func(_ string, _ access_types.PluginAccessAction, _ []byte) error {
				listener.Send(plugin_entities.SessionMessage{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
					Data: []byte(fmt.Sprintf(`{"type":"blob_chunk","message":{"id":"file","total_length":%s,"blob":"","end":false}}`, length)),
				})
				listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_END})
				return nil
			}
			session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_AGENT_STRATEGY,
				access_types.PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY)
			response, err := InvokeAgentStrategy(session, &requests.RequestInvokeAgentStrategy{})
			require.NoError(t, err)
			defer response.Close()
			require.True(t, response.Next())
			_, err = response.Read()
			require.EqualError(t, err, "invalid file length")
		})
	}
}
