package model_entities

import "encoding/json"

type PollingStatus string

type ModelPollingResult struct {
	Status                PollingStatus   `json:"status" validate:"required,oneof=running succeeded failed"`
	PluginState           map[string]any  `json:"plugin_state,omitempty"`
	Result                json.RawMessage `json:"result,omitempty"`
	Error                 string          `json:"error,omitempty"`
	NextCheckAfterSeconds *int            `json:"next_check_after_seconds,omitempty"`
	ExpiresAfterSeconds   *int            `json:"expires_after_seconds,omitempty"`
	MaxAttempts           *int            `json:"max_attempts,omitempty"`
}
