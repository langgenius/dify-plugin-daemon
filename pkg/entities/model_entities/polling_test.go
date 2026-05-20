package model_entities

import (
	"encoding/json"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

func TestModelPollingResultValidatesStatus(t *testing.T) {
	nextCheckAfterSeconds := 10
	result := ModelPollingResult{
		Status:                PollingStatus("running"),
		PluginState:           map[string]any{"task_id": "task-1"},
		NextCheckAfterSeconds: &nextCheckAfterSeconds,
	}

	if err := validators.GlobalEntitiesValidator.Struct(result); err != nil {
		t.Fatalf("validate polling result: %v", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal polling result: %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("invalid json: %s", data)
	}
}

func TestModelPollingResultRejectsUnknownStatus(t *testing.T) {
	result := ModelPollingResult{
		Status: PollingStatus("pending"),
	}

	if err := validators.GlobalEntitiesValidator.Struct(result); err == nil {
		t.Fatal("expected unknown polling status validation error")
	}
}

func TestModelPollingResultRequiresRunningState(t *testing.T) {
	result := ModelPollingResult{
		Status: PollingStatus("running"),
	}

	if err := validators.GlobalEntitiesValidator.Struct(result); err == nil {
		t.Fatal("expected missing plugin_state validation error")
	}
}

func TestModelPollingResultRequiresSucceededResult(t *testing.T) {
	result := ModelPollingResult{
		Status: PollingStatus("succeeded"),
		Result: json.RawMessage("null"),
	}

	if err := validators.GlobalEntitiesValidator.Struct(result); err == nil {
		t.Fatal("expected missing result validation error")
	}
}

func TestModelPollingResultRequiresFailedError(t *testing.T) {
	result := ModelPollingResult{
		Status: PollingStatus("failed"),
		Error:  "  ",
	}

	if err := validators.GlobalEntitiesValidator.Struct(result); err == nil {
		t.Fatal("expected missing error validation error")
	}
}

func TestModelPollingResultAcceptsTerminalPayloads(t *testing.T) {
	succeeded := ModelPollingResult{
		Status: PollingStatus("succeeded"),
		Result: json.RawMessage(`{"text":"done"}`),
	}
	if err := validators.GlobalEntitiesValidator.Struct(succeeded); err != nil {
		t.Fatalf("validate succeeded result: %v", err)
	}

	failed := ModelPollingResult{
		Status: PollingStatus("failed"),
		Error:  "provider failed",
	}
	if err := validators.GlobalEntitiesValidator.Struct(failed); err != nil {
		t.Fatalf("validate failed result: %v", err)
	}
}
