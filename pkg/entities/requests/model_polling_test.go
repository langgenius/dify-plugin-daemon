package requests

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
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

	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			t.Fatalf("unexpected validation error type: %T: %v", err, err)
		}
		if len(validationErrors) != 1 {
			t.Fatalf("expected 1 validation error, got %d: %v", len(validationErrors), validationErrors)
		}
		validationError := validationErrors[0]
		if validationError.Field() != "PluginState" {
			t.Fatalf("unexpected field: %s", validationError.Field())
		}
		if validationError.Tag() != "required" {
			t.Fatalf("unexpected tag: %s", validationError.Tag())
		}
		return
	}
	t.Fatal("expected missing plugin_state validation error")
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

	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			t.Fatalf("unexpected validation error type: %T: %v", err, err)
		}
		if len(validationErrors) != 1 {
			t.Fatalf("expected 1 validation error, got %d: %v", len(validationErrors), validationErrors)
		}
		validationError := validationErrors[0]
		if validationError.Field() != "PluginState" {
			t.Fatalf("unexpected field: %s", validationError.Field())
		}
		if validationError.Tag() != "min" {
			t.Fatalf("unexpected tag: %s", validationError.Tag())
		}
		return
	}
	t.Fatal("expected empty plugin_state validation error")
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

func TestPollingRequestMapExcludesSuspensionFields(t *testing.T) {
	startPolling := RequestStartPolling{
		RequestInvokeLLM: RequestInvokeLLM{
			BaseRequestInvokeModel: BaseRequestInvokeModel{
				Provider: "volcengine",
				Model:    "doubao-seedance-2-0-260128",
			},
			Credentials: Credentials{
				Credentials: map[string]any{"ark_api_key": "test"},
			},
			InvokeLLMSchema: InvokeLLMSchema{
				Stream: true,
			},
			ModelType: model_entities.MODEL_TYPE_LLM,
		},
	}

	assertPollingRequestMapShape(t, parser.StructToMap(startPolling), false)

	checkPolling := RequestCheckPolling{
		BaseRequestInvokeModel: BaseRequestInvokeModel{
			Provider: "volcengine",
			Model:    "doubao-seedance-2-0-260128",
		},
		Credentials: Credentials{
			Credentials: map[string]any{"ark_api_key": "test"},
		},
		ModelType:   model_entities.MODEL_TYPE_LLM,
		PluginState: map[string]any{"task_id": "task-1"},
	}

	assertPollingRequestMapShape(t, parser.StructToMap(checkPolling), true)
}

func assertPollingRequestMapShape(t *testing.T, result map[string]any, expectPluginState bool) {
	t.Helper()

	if _, ok := result["workflow_run_id"]; ok {
		t.Fatal("workflow_run_id should not be present in polling request map")
	}
	if _, ok := result["node_id"]; ok {
		t.Fatal("node_id should not be present in polling request map")
	}

	if result["provider"] != "volcengine" {
		t.Fatalf("provider should be volcengine, got %#v", result["provider"])
	}
	if result["model"] != "doubao-seedance-2-0-260128" {
		t.Fatalf("model should be present, got %#v", result["model"])
	}
	if result["model_type"] != model_entities.MODEL_TYPE_LLM {
		t.Fatalf("model_type should be llm, got %#v", result["model_type"])
	}

	credentials, ok := result["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("credentials should be a map, got %T", result["credentials"])
	}
	if credentials["ark_api_key"] != "test" {
		t.Fatalf("credentials should contain ark_api_key, got %#v", credentials)
	}

	if expectPluginState {
		pluginState, ok := result["plugin_state"].(map[string]any)
		if !ok {
			t.Fatalf("plugin_state should be a map, got %T", result["plugin_state"])
		}
		if pluginState["task_id"] != "task-1" {
			t.Fatalf("plugin_state should contain task_id, got %#v", pluginState)
		}
		return
	}

	if _, ok := result["plugin_state"]; ok {
		t.Fatal("plugin_state should not be present in start polling request map")
	}
	if stream, ok := result["stream"].(bool); !ok || !stream {
		t.Fatalf("stream should be true, got %#v", result["stream"])
	}
}
