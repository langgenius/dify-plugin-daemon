package plugin_entities

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParsePluginUniversalEvent_RecoversFromSessionHandlerPanic(t *testing.T) {
	event := PluginUniversalEvent{
		SessionId: "session-1",
		Event:     PLUGIN_EVENT_SESSION,
		Data:      json.RawMessage(`{"type":"invoke","data":{"x":1}}`),
	}

	var gotErr string

	ParsePluginUniversalEvent(
		mustMarshalEvent(t, event),
		"",
		func(string, []byte) {
			panic("boom")
		},
		nil,
		func(err string) {
			gotErr = err
		},
		nil,
	)

	if !strings.Contains(gotErr, "plugin event handler panic: boom") {
		t.Fatalf("expected panic to be forwarded to error handler, got %q", gotErr)
	}
}

func TestParsePluginUniversalEvent_NilHandlersDoNotPanic(t *testing.T) {
	cases := []PluginUniversalEvent{
		{Event: PLUGIN_EVENT_SESSION, Data: json.RawMessage(`{"type":"invoke","data":{}}`)},
		{Event: PLUGIN_EVENT_ERROR, Data: json.RawMessage(`"err"`)},
		{Event: PLUGIN_EVENT_HEARTBEAT, Data: json.RawMessage(`null`)},
	}

	for _, event := range cases {
		ParsePluginUniversalEvent(mustMarshalEvent(t, event), "", nil, nil, nil, nil)
	}
}

func mustMarshalEvent(t *testing.T, event PluginUniversalEvent) []byte {
	t.Helper()

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	return data
}
