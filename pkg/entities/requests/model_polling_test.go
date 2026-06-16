package requests

import (
	"encoding/json"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

func TestRequestStartPollingDoesNotRequireSuspensionFields(t *testing.T) {
	payload := []byte(`{
		"provider": "volcengine",
		"model": "doubao-seedance-2-0-260128",
		"model_type": "llm",
		"credentials": {"ark_api_key": "test"}
	}`)

	var request RequestStartPolling
	if err := json.Unmarshal(payload, &request); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		t.Fatalf("validate request: %v", err)
	}
}

func TestRequestCheckPollingRequiresPluginState(t *testing.T) {
	payload := []byte(`{
		"provider": "volcengine",
		"model": "doubao-seedance-2-0-260128",
		"model_type": "llm",
		"credentials": {"ark_api_key": "test"}
	}`)

	var request RequestCheckPolling
	if err := json.Unmarshal(payload, &request); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if err := validators.GlobalEntitiesValidator.Struct(request); err == nil {
		t.Fatal("expected missing plugin_state validation error")
	}
}

func TestRequestCheckPollingRejectsEmptyPluginState(t *testing.T) {
	payload := []byte(`{
		"provider": "volcengine",
		"model": "doubao-seedance-2-0-260128",
		"model_type": "llm",
		"credentials": {"ark_api_key": "test"},
		"plugin_state": {}
	}`)

	var request RequestCheckPolling
	if err := json.Unmarshal(payload, &request); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if err := validators.GlobalEntitiesValidator.Struct(request); err == nil {
		t.Fatal("expected empty plugin_state validation error")
	}
}

func TestRequestCheckPollingAcceptsPluginState(t *testing.T) {
	payload := []byte(`{
		"provider": "volcengine",
		"model": "doubao-seedance-2-0-260128",
		"model_type": "llm",
		"credentials": {"ark_api_key": "test"},
		"plugin_state": {"task_id": "task-1"}
	}`)

	var request RequestCheckPolling
	if err := json.Unmarshal(payload, &request); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		t.Fatalf("validate request: %v", err)
	}
	if request.ModelType != model_entities.MODEL_TYPE_LLM {
		t.Fatalf("unexpected model_type: %s", request.ModelType)
	}
	if request.PluginState["task_id"] != "task-1" {
		t.Fatalf("unexpected plugin_state: %#v", request.PluginState)
	}
}
