package requests

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
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
