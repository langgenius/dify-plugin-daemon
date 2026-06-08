package session_manager

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionMessagePayload(t *testing.T) {
	conversationID := "conversation-id"
	messageID := "message-id"
	appID := "app-id"
	endpointID := "endpoint-id"

	session := NewSession(NewSessionPayload{
		ConversationID: &conversationID,
		MessageID:      &messageID,
		AppID:          &appID,
		EndpointID:     &endpointID,
		Context: map[string]any{
			"trace": "trace-value",
		},
		IgnoreCache: true,
	})
	t.Cleanup(func() {
		DeleteSession(DeleteSessionPayload{ID: session.ID, IgnoreCache: true})
	})

	payload := session.Message(PLUGIN_IN_STREAM_EVENT_REQUEST, map[string]any{
		"input": "hello",
	})

	var got map[string]any
	require.NoError(t, json.Unmarshal(payload, &got))
	require.Equal(t, session.ID, got["session_id"])
	require.Equal(t, conversationID, got["conversation_id"])
	require.Equal(t, messageID, got["message_id"])
	require.Equal(t, appID, got["app_id"])
	require.Equal(t, endpointID, got["endpoint_id"])
	require.Equal(t, string(PLUGIN_IN_STREAM_EVENT_REQUEST), got["event"])
	require.Equal(t, map[string]any{"trace": "trace-value"}, got["context"])
	require.Equal(t, map[string]any{"input": "hello"}, got["data"])
}
