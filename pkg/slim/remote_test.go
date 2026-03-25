package slim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLookupRoute_KnownActions(t *testing.T) {
	tests := []struct {
		action   string
		wantType string
		wantPath string
	}{
		{"invoke_tool", "tool", "/tool/invoke"},
		{"invoke_llm", "model", "/llm/invoke"},
		{"invoke_agent_strategy", "agent_strategy", "/agent_strategy/invoke"},
		{"invoke_endpoint", "endpoint", "/endpoint/invoke"},
		{"get_authorization_url", "oauth", "/oauth/get_authorization_url"},
		{"validate_datasource_credentials", "datasource", "/datasource/validate_credentials"},
		{"fetch_parameter_options", "dynamic_parameter", "/dynamic_select/fetch_parameter_options"},
		{"invoke_trigger_event", "trigger", "/trigger/invoke_event"},
	}
	for _, tt := range tests {
		route, ok := LookupRoute(tt.action)
		if !ok {
			t.Errorf("LookupRoute(%q) not found", tt.action)
			continue
		}
		if route.Type != tt.wantType {
			t.Errorf("LookupRoute(%q).Type = %q; want %q", tt.action, route.Type, tt.wantType)
		}
		if route.Path != tt.wantPath {
			t.Errorf("LookupRoute(%q).Path = %q; want %q", tt.action, route.Path, tt.wantPath)
		}
	}
}

func TestLookupRoute_UnknownAction(t *testing.T) {
	_, ok := LookupRoute("nonexistent")
	if ok {
		t.Error("LookupRoute(nonexistent) should return false")
	}
}

func TestDaemonClient_Dispatch_UnknownAction(t *testing.T) {
	client := NewDaemonClient("http://localhost:9999", "key")
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "bad_action",
		Request:  RequestMeta{Data: json.RawMessage(`{}`)},
	}
	_, err := client.Dispatch(ctx)
	if err == nil {
		t.Fatal("Dispatch() should fail for unknown action")
	}
	se := err.(*SlimError)
	if se.Code != ErrUnknownAction {
		t.Errorf("Code = %q; want %q", se.Code, ErrUnknownAction)
	}
}

func TestDaemonClient_Dispatch_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()

	client := NewDaemonClient(srv.URL, "testkey")
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}
	_, err := client.Dispatch(ctx)
	if err == nil {
		t.Fatal("Dispatch() should fail on non-200 status")
	}
	se := err.(*SlimError)
	if se.Code != ErrDaemon {
		t.Errorf("Code = %q; want %q", se.Code, ErrDaemon)
	}
	if !strings.Contains(se.Message, "500") {
		t.Errorf("Message should contain status code 500: %q", se.Message)
	}
}

func TestDaemonClient_Dispatch_SetsHeaders(t *testing.T) {
	var gotHeaders http.Header
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewDaemonClient(srv.URL, "my-api-key")
	ctx := &InvokeContext{
		PluginID: "author/plugin:1.0.0",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}
	body, err := client.Dispatch(ctx)
	if err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}
	body.Close()

	if gotHeaders.Get("X-Api-Key") != "my-api-key" {
		t.Errorf("X-Api-Key = %q; want %q", gotHeaders.Get("X-Api-Key"), "my-api-key")
	}
	if gotHeaders.Get("X-Plugin-Unique-Identifier") != "author/plugin:1.0.0" {
		t.Errorf("X-Plugin-Unique-Identifier = %q; want %q",
			gotHeaders.Get("X-Plugin-Unique-Identifier"), "author/plugin:1.0.0")
	}
	if gotHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q; want %q", gotHeaders.Get("Content-Type"), "application/json")
	}
	wantPath := dispatchPrefix + "/tool/invoke"
	if gotPath != wantPath {
		t.Errorf("path = %q; want %q", gotPath, wantPath)
	}
}

func TestRunRemote_SSEStream(t *testing.T) {
	chunks := []map[string]any{
		{"result": "chunk1"},
		{"result": "chunk2"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		for _, chunk := range chunks {
			b, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", b)
		}
	}))
	defer srv.Close()

	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	remote := &RemoteConfig{DaemonAddr: srv.URL, DaemonKey: "key"}
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}

	if err := RunRemote(ctx, remote, out); err != nil {
		t.Fatalf("RunRemote() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d output lines; want 3 (2 chunks + done)", len(lines))
	}

	for i := 0; i < 2; i++ {
		var evt outputEvent
		if err := json.Unmarshal([]byte(lines[i]), &evt); err != nil {
			t.Fatalf("line %d: json.Unmarshal() error: %v", i, err)
		}
		if evt.Event != "chunk" {
			t.Errorf("line %d: event = %q; want %q", i, evt.Event, "chunk")
		}
	}

	var doneEvt outputEvent
	if err := json.Unmarshal([]byte(lines[2]), &doneEvt); err != nil {
		t.Fatalf("done line: json.Unmarshal() error: %v", err)
	}
	if doneEvt.Event != "done" {
		t.Errorf("last event = %q; want %q", doneEvt.Event, "done")
	}
}

func TestRunRemote_InvalidSSEJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {not-valid-json}\n\n")
	}))
	defer srv.Close()

	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	remote := &RemoteConfig{DaemonAddr: srv.URL, DaemonKey: "key"}
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}

	err := RunRemote(ctx, remote, out)
	if err == nil {
		t.Fatal("RunRemote() should fail on invalid SSE JSON")
	}
	se := err.(*SlimError)
	if se.Code != ErrStreamParse {
		t.Errorf("Code = %q; want %q", se.Code, ErrStreamParse)
	}
}

func TestRunRemote_SkipsNonDataLines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "event: ping\n")
		fmt.Fprint(w, ": comment\n")
		fmt.Fprint(w, "data: {\"ok\":true}\n\n")
	}))
	defer srv.Close()

	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	remote := &RemoteConfig{DaemonAddr: srv.URL, DaemonKey: "key"}
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}

	if err := RunRemote(ctx, remote, out); err != nil {
		t.Fatalf("RunRemote() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d output lines; want 2 (1 chunk + done)", len(lines))
	}
}

func TestRunRemote_DaemonError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	remote := &RemoteConfig{DaemonAddr: srv.URL, DaemonKey: "key"}
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}

	err := RunRemote(ctx, remote, out)
	if err == nil {
		t.Fatal("RunRemote() should fail on daemon error")
	}
	se := err.(*SlimError)
	if se.Code != ErrDaemon {
		t.Errorf("Code = %q; want %q", se.Code, ErrDaemon)
	}
}
