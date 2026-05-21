package model_entities

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type PollingStatus string

const (
	PollingStatusRunning   PollingStatus = "running"
	PollingStatusSucceeded PollingStatus = "succeeded"
	PollingStatusFailed    PollingStatus = "failed"
)

type ModelPollingResult struct {
	Status                PollingStatus   `json:"status" validate:"required,oneof=running succeeded failed"`
	PluginState           map[string]any  `json:"plugin_state,omitempty"`
	Result                json.RawMessage `json:"result,omitempty"`
	Error                 string          `json:"error,omitempty"`
	NextCheckAfterSeconds *int            `json:"next_check_after_seconds,omitempty"`
	ExpiresAfterSeconds   *int            `json:"expires_after_seconds,omitempty"`
	MaxAttempts           *int            `json:"max_attempts,omitempty"`
}

func init() {
	validators.GlobalEntitiesValidator.RegisterStructValidation(
		validateModelPollingResult,
		ModelPollingResult{},
	)
}

func validateModelPollingResult(sl validator.StructLevel) {
	result, ok := sl.Current().Interface().(ModelPollingResult)
	if !ok {
		return
	}

	switch result.Status {
	case PollingStatusRunning:
		if len(result.PluginState) == 0 {
			sl.ReportError(result.PluginState, "plugin_state", "PluginState", "required_for_running", "")
		}
	case PollingStatusSucceeded:
		if isEmptyJSON(result.Result) {
			sl.ReportError(result.Result, "result", "Result", "required_for_succeeded", "")
		}
	case PollingStatusFailed:
		if strings.TrimSpace(result.Error) == "" {
			sl.ReportError(result.Error, "error", "Error", "required_for_failed", "")
		}
	}
}

func isEmptyJSON(data json.RawMessage) bool {
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}
