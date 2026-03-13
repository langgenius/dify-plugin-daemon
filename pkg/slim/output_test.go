package slim

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOutputWriter_Chunk(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	if err := out.Chunk(json.RawMessage(`{"result":"ok"}`)); err != nil {
		t.Fatalf("Chunk() error: %v", err)
	}
	var evt outputEvent
	if err := json.Unmarshal(buf.Bytes(), &evt); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if evt.Event != "chunk" {
		t.Errorf("Event = %q; want %q", evt.Event, "chunk")
	}
}

func TestOutputWriter_Done(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	if err := out.Done(); err != nil {
		t.Fatalf("Done() error: %v", err)
	}
	var evt outputEvent
	if err := json.Unmarshal(buf.Bytes(), &evt); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if evt.Event != "done" {
		t.Errorf("Event = %q; want %q", evt.Event, "done")
	}
}

func TestOutputWriter_Error(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	if err := out.Error(ErrPluginExec, "plugin crashed"); err != nil {
		t.Fatalf("Error() error: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if raw["event"] != "error" {
		t.Errorf("event = %v; want %q", raw["event"], "error")
	}
	data, ok := raw["data"].(map[string]any)
	if !ok {
		t.Fatalf("data is not a map: %T", raw["data"])
	}
	if data["code"] != string(ErrPluginExec) {
		t.Errorf("data.code = %v; want %q", data["code"], ErrPluginExec)
	}
	if data["message"] != "plugin crashed" {
		t.Errorf("data.message = %v; want %q", data["message"], "plugin crashed")
	}
}

func TestOutputWriter_Message(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	if err := out.Message("init", "starting up"); err != nil {
		t.Fatalf("Message() error: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if raw["event"] != "message" {
		t.Errorf("event = %v; want %q", raw["event"], "message")
	}
	data := raw["data"].(map[string]any)
	if data["stage"] != "init" {
		t.Errorf("stage = %v; want %q", data["stage"], "init")
	}
	if data["message"] != "starting up" {
		t.Errorf("message = %v; want %q", data["message"], "starting up")
	}
}

func TestOutputWriter_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	out := NewOutputWriter(&buf)
	out.Chunk(json.RawMessage(`{"a":1}`))
	out.Chunk(json.RawMessage(`{"a":2}`))
	out.Done()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines; want 3", len(lines))
	}
	for i, line := range lines {
		var evt outputEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("line %d: json.Unmarshal() error: %v", i, err)
		}
		if i < 2 && evt.Event != "chunk" {
			t.Errorf("line %d: event = %q; want %q", i, evt.Event, "chunk")
		}
		if i == 2 && evt.Event != "done" {
			t.Errorf("line %d: event = %q; want %q", i, evt.Event, "done")
		}
	}
}
