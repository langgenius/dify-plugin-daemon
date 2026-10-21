package io_tunnel

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func TestInvokeEndpointDrainsBufferedResponse(t *testing.T) {
	routine.InitPool(2)
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete", true: "plugin error"}[failure], func(t *testing.T) {
			listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
			runtime := &fakePluginRuntime{
				listener:         listener,
				uniqueIdentifier: testPluginUniqueIdentifier,
				runtimeType:      plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
			}
			runtime.writeFn = func(string, access_types.PluginAccessAction, []byte) error {
				listener.Send(plugin_entities.SessionMessage{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
					Data: []byte(`{"status":201,"headers":{"Content-Type":"text/plain"},"result":"707265666978"}`),
				})
				listener.Send(plugin_entities.SessionMessage{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM,
					Data: []byte(`{"result":"737566666978"}`),
				})
				if failure {
					listener.Send(plugin_entities.SessionMessage{
						Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
						Data: mustMarshalPluginError(t, plugin_entities.ErrorResponse{
							ErrorType: "RuntimeError", Message: "cannot open /private/example",
						}),
					})
				} else {
					listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_END})
				}
				return nil
			}
			session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, access_types.PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT)
			status, headers, response, err := InvokeEndpoint(session, &requests.RequestInvokeEndpoint{})
			require.NoError(t, err)
			defer response.Close()
			require.Equal(t, http.StatusCreated, status)
			require.Equal(t, "text/plain", headers.Get("Content-Type"))
			var body []byte
			for response.Next() {
				var chunk []byte
				chunk, err = response.Read()
				if err != nil {
					break
				}
				body = append(body, chunk...)
			}
			require.Equal(t, "prefixsuffix", string(body))
			if failure {
				require.ErrorContains(t, err, "/private/example")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
