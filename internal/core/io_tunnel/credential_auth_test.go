package io_tunnel

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func TestCredentialAuthDispatch(t *testing.T) {
	for _, runtimeType := range []plugin_entities.PluginRuntimeType{
		plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
		plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
	} {
		for _, tt := range []struct {
			name     string
			action   access_types.PluginAccessAction
			response string
		}{
			{"detailed", access_types.PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS, `{"status":"valid","reason_code":null,"retryable":false}`},
			{"old sdk", access_types.PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS, `{"result":true}`},
			{"revoke", access_types.PLUGIN_ACCESS_ACTION_REVOKE_CREDENTIALS, `{"status":"revoked","reason_code":null,"retryable":false}`},
		} {
			t.Run(string(runtimeType)+"/"+tt.name, func(t *testing.T) {
				listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
				runtime := &fakePluginRuntime{
					listener: listener, uniqueIdentifier: testPluginUniqueIdentifier, runtimeType: runtimeType,
				}
				var received struct {
					Data map[string]any `json:"data"`
				}
				runtime.writeFn = func(_ string, action access_types.PluginAccessAction, data []byte) error {
					require.Equal(t, tt.action, action)
					require.NoError(t, json.Unmarshal(data, &received))
					listener.Send(plugin_entities.SessionMessage{
						Type: plugin_entities.SESSION_MESSAGE_TYPE_STREAM, Data: []byte(tt.response),
					})
					listener.Send(plugin_entities.SessionMessage{Type: plugin_entities.SESSION_MESSAGE_TYPE_END})
					return nil
				}
				accessType := access_types.PLUGIN_ACCESS_TYPE_TOOL
				if tt.action == access_types.PLUGIN_ACCESS_ACTION_REVOKE_CREDENTIALS {
					accessType = access_types.PLUGIN_ACCESS_TYPE_OAUTH
				}
				session := newTestSession(t, runtime, accessType, tt.action)
				credentials := map[string]any{"access_token": "test-token"}
				var encoded []byte
				if tt.action == access_types.PLUGIN_ACCESS_ACTION_REVOKE_CREDENTIALS {
					response, err := RevokeCredentials(session, &requests.RequestOAuthRevokeCredentials{
						Provider: "calendar", Credentials: credentials,
						SystemCredentials: map[string]any{"client_secret": "test-client-secret"},
					})
					require.NoError(t, err)
					defer response.Close()
					require.True(t, response.Next())
					result, err := response.Read()
					require.NoError(t, err)
					encoded, err = json.Marshal(result)
					require.NoError(t, err)
					require.Equal(t, map[string]any{"client_secret": "test-client-secret"}, received.Data["system_credentials"])
				} else {
					response, err := ValidateToolCredentials(session, &requests.RequestValidateToolCredentials{
						Provider: "calendar", Credentials: credentials, CredentialType: "oauth2", ResultFormat: "detailed",
					})
					require.NoError(t, err)
					defer response.Close()
					require.True(t, response.Next())
					result, err := response.Read()
					require.NoError(t, err)
					encoded, err = json.Marshal(result)
					require.NoError(t, err)
					require.Equal(t, "oauth2", received.Data["credential_type"])
					require.Equal(t, "detailed", received.Data["result_format"])
				}
				require.JSONEq(t, tt.response, string(encoded))
				require.Equal(t, "user-1", received.Data["user_id"])
				require.Equal(t, "calendar", received.Data["provider"])
				require.Equal(t, credentials, received.Data["credentials"])
			})
		}
	}
}
