package slim

import (
	"encoding/json"
	"testing"
)

func TestTransformRequest_ValidToolAction(t *testing.T) {
	ctx := &InvokeContext{
		PluginID: "author/plugin:1.0.0",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "tenant-1",
			UserID:   "user-1",
			Data:     json.RawMessage(`{"tool_name":"search","params":{"q":"test"}}`),
		},
	}
	b, sessionID, err := TransformRequest(ctx)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}
	if sessionID == "" {
		t.Fatal("sessionID should not be empty")
	}

	var msg map[string]any
	if err := json.Unmarshal(b, &msg); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if msg["tenant_id"] != "tenant-1" {
		t.Errorf("tenant_id = %v; want %q", msg["tenant_id"], "tenant-1")
	}
	if msg["session_id"] != sessionID {
		t.Errorf("session_id = %v; want %q", msg["session_id"], sessionID)
	}
	if msg["event"] != "request" {
		t.Errorf("event = %v; want %q", msg["event"], "request")
	}

	data, ok := msg["data"].(map[string]any)
	if !ok {
		t.Fatalf("data is not a map: %T", msg["data"])
	}
	if data["type"] != "tool" {
		t.Errorf("data.type = %v; want %q", data["type"], "tool")
	}
	if data["action"] != "invoke_tool" {
		t.Errorf("data.action = %v; want %q", data["action"], "invoke_tool")
	}
	if data["user_id"] != "user-1" {
		t.Errorf("data.user_id = %v; want %q", data["user_id"], "user-1")
	}
}

func TestTransformRequest_ModelAction(t *testing.T) {
	ctx := &InvokeContext{
		PluginID: "author/plugin:1.0.0",
		Action:   "invoke_llm",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{"model":"gpt-4"}`),
		},
	}
	b, _, err := TransformRequest(ctx)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}
	var msg map[string]any
	if err := json.Unmarshal(b, &msg); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	data := msg["data"].(map[string]any)
	if data["type"] != "model" {
		t.Errorf("data.type = %v; want %q", data["type"], "model")
	}
	if data["action"] != "invoke_llm" {
		t.Errorf("data.action = %v; want %q", data["action"], "invoke_llm")
	}
}

func TestTransformRequest_UnknownAction(t *testing.T) {
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "nonexistent_action",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{}`),
		},
	}
	_, _, err := TransformRequest(ctx)
	if err == nil {
		t.Fatal("TransformRequest() should fail for unknown action")
	}
	se, ok := err.(*SlimError)
	if !ok {
		t.Fatalf("expected *SlimError, got %T", err)
	}
	if se.Code != ErrUnknownAction {
		t.Errorf("Code = %q; want %q", se.Code, ErrUnknownAction)
	}
}

func TestTransformRequest_InvalidDataJSON(t *testing.T) {
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`not-json`),
		},
	}
	_, _, err := TransformRequest(ctx)
	if err == nil {
		t.Fatal("TransformRequest() should fail for invalid data JSON")
	}
	se, ok := err.(*SlimError)
	if !ok {
		t.Fatalf("expected *SlimError, got %T", err)
	}
	if se.Code != ErrInvalidArgsJSON {
		t.Errorf("Code = %q; want %q", se.Code, ErrInvalidArgsJSON)
	}
}

func TestTransformRequest_MarshalErrorReturnsSlimError(t *testing.T) {
	ctx := &InvokeContext{
		PluginID: "plugin",
		Action:   "invoke_tool",
		Request: RequestMeta{
			TenantID: "t1",
			Data:     json.RawMessage(`{"key":"val"}`),
		},
	}
	_, _, err := TransformRequest(ctx)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}
}

func TestTransformRequest_AllActionRoutes(t *testing.T) {
	for action, route := range ActionRoutes {
		ctx := &InvokeContext{
			PluginID: "plugin",
			Action:   action,
			Request: RequestMeta{
				TenantID: "t1",
				Data:     json.RawMessage(`{}`),
			},
		}
		b, _, err := TransformRequest(ctx)
		if err != nil {
			t.Errorf("TransformRequest(%q) error: %v", action, err)
			continue
		}
		var msg map[string]any
		if err := json.Unmarshal(b, &msg); err != nil {
			t.Errorf("json.Unmarshal for action %q error: %v", action, err)
			continue
		}
		data := msg["data"].(map[string]any)
		if data["type"] != route.Type {
			t.Errorf("action %q: data.type = %v; want %q", action, data["type"], route.Type)
		}
		if data["action"] != action {
			t.Errorf("action %q: data.action = %v; want %q", action, data["action"], action)
		}
	}
}
