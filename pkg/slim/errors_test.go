package slim

import (
	"encoding/json"
	"testing"
)

func TestSlimError_Error(t *testing.T) {
	err := NewError(ErrInvalidInput, "bad input")
	if err.Error() != "INVALID_INPUT: bad input" {
		t.Fatalf("Error() = %q; want %q", err.Error(), "INVALID_INPUT: bad input")
	}
}

func TestSlimError_ExitCode(t *testing.T) {
	tests := []struct {
		code ErrorCode
		want int
	}{
		{ErrInvalidInput, ExitInputError},
		{ErrInvalidArgsJSON, ExitInputError},
		{ErrConfigLoad, ExitInputError},
		{ErrConfigInvalid, ExitInputError},
		{ErrUnknownMode, ExitInputError},
		{ErrUnknownAction, ExitInputError},
		{ErrNetwork, ExitNetworkError},
		{ErrPluginDownload, ExitNetworkError},
		{ErrPluginDownloadTimeout, ExitNetworkError},
		{ErrDaemon, ExitDaemonError},
		{ErrPluginExec, ExitPluginError},
		{ErrStreamParse, ExitPluginError},
		{ErrPluginInit, ExitPluginError},
		{ErrPluginNotFound, ExitPluginError},
	}
	for _, tt := range tests {
		err := NewError(tt.code, "msg")
		if got := err.ExitCode(); got != tt.want {
			t.Errorf("ExitCode() for %s = %d; want %d", tt.code, got, tt.want)
		}
	}
}

func TestSlimError_JSON(t *testing.T) {
	err := NewError(ErrNetwork, "connection refused")
	b, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("json.Marshal() error: %v", jsonErr)
	}
	var decoded SlimError
	if jsonErr := json.Unmarshal(b, &decoded); jsonErr != nil {
		t.Fatalf("json.Unmarshal() error: %v", jsonErr)
	}
	if decoded.Code != ErrNetwork {
		t.Errorf("Code = %q; want %q", decoded.Code, ErrNetwork)
	}
	if decoded.Message != "connection refused" {
		t.Errorf("Message = %q; want %q", decoded.Message, "connection refused")
	}
}
