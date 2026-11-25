package io_tunnel

import (
	"encoding/json"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

func TestRawModelDispatchReachesRuntimeWithPreciseJSON(t *testing.T) {
	for _, runtimeType := range []plugin_entities.PluginRuntimeType{
		plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
	} {
		t.Run(string(runtimeType), func(t *testing.T) {
			listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
			runtime := &fakePluginRuntime{listener: listener, uniqueIdentifier: testPluginUniqueIdentifier, runtimeType: runtimeType}
			action := access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM
			session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_MODEL, action)
			request := ModelDispatchData{
				"provider":         json.RawMessage(`"target"`),
				"model":            json.RawMessage(`"model"`),
				"model_parameters": json.RawMessage(`{"future":[9007199254740993,1.00000000000000000003]}`),
				"prompt_messages":  json.RawMessage(`[{"role":"user","future_media":{"data":"AA+/"}}]`),
				"extra":            json.RawMessage(`null`),
				"action":           json.RawMessage(`"not-a-trusted-action"`),
				"type":             json.RawMessage(`"not-a-trusted-type"`),
				"user_id":          json.RawMessage(`"not-a-trusted-user"`),
			}
			runtime.writeFn = func(_ string, gotAction access_types.PluginAccessAction, data []byte) error {
				require.Equal(t, action, gotAction)
				var frame struct {
					Data ModelDispatchData `json:"data"`
				}
				require.NoError(t, json.Unmarshal(data, &frame))
				for _, key := range []string{"provider", "model", "model_parameters", "prompt_messages", "extra"} {
					require.Equal(t, request[key], frame.Data[key], key)
				}
				require.Equal(t, `"invoke_llm"`, string(frame.Data["action"]))
				require.Equal(t, `"model"`, string(frame.Data["type"]))
				require.Equal(t, `"user-1"`, string(frame.Data["user_id"]))
				for _, chunk := range []string{`"first"`, `"second"`} {
					listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM, Data: []byte(chunk)})
				}
				listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_END})
				return nil
			}
			response, err := GenericInvokePlugin[ModelDispatchData, string](session, &request, 2)
			require.NoError(t, err)
			defer response.Close()
			var chunks []string
			for response.Next() {
				chunk, err := response.Read()
				require.NoError(t, err)
				chunks = append(chunks, chunk)
			}
			require.Equal(t, []string{"first", "second"}, chunks)
		})
	}
}
