package app

import (
	"testing"
)

func TestServerHostDefault(t *testing.T) {
	tests := []struct {
		name         string
		inputHost    string
		expectedHost string
	}{
		{
			name:         "empty host should default to 0.0.0.0",
			inputHost:    "",
			expectedHost: "0.0.0.0",
		},
		{
			name:         "custom host should be preserved",
			inputHost:    "127.0.0.1",
			expectedHost: "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ServerHost: tt.inputHost,
				ServerPort: 5002,
				ServerKey:  "test-key",
			}
			config.SetDefault()

			if config.ServerHost != tt.expectedHost {
				t.Errorf("expected ServerHost %s, got %s", tt.expectedHost, config.ServerHost)
			}
		})
	}
}
