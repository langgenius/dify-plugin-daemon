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

// TestProxyEnvVarFallback verifies that SetDefault falls back to lowercase
// proxy environment variables (http_proxy, https_proxy, no_proxy) when the
// uppercase equivalents were not populated by envconfig. This mirrors the
// behaviour already supported by the sandbox and api containers and is the
// subject of langgenius/dify#18752.
func TestProxyEnvVarFallback(t *testing.T) {
	t.Run("lowercase proxy vars fill empty fields", func(t *testing.T) {
		t.Setenv("http_proxy", "http://lowercase-host:8080")
		t.Setenv("https_proxy", "http://lowercase-host:8443")
		t.Setenv("no_proxy", "localhost,127.0.0.1")

		config := &Config{}
		config.SetDefault()

		if config.HttpProxy != "http://lowercase-host:8080" {
			t.Errorf("expected HttpProxy from http_proxy, got %q", config.HttpProxy)
		}
		if config.HttpsProxy != "http://lowercase-host:8443" {
			t.Errorf("expected HttpsProxy from https_proxy, got %q", config.HttpsProxy)
		}
		if config.NoProxy != "localhost,127.0.0.1" {
			t.Errorf("expected NoProxy from no_proxy, got %q", config.NoProxy)
		}
	})

	t.Run("uppercase values take precedence over lowercase", func(t *testing.T) {
		t.Setenv("http_proxy", "http://lowercase-host:8080")
		t.Setenv("https_proxy", "http://lowercase-host:8443")
		t.Setenv("no_proxy", "localhost,127.0.0.1")

		config := &Config{
			HttpProxy:  "http://uppercase-host:8080",
			HttpsProxy: "http://uppercase-host:8443",
			NoProxy:    "10.0.0.0/8",
		}
		config.SetDefault()

		if config.HttpProxy != "http://uppercase-host:8080" {
			t.Errorf("uppercase HttpProxy should be preserved, got %q", config.HttpProxy)
		}
		if config.HttpsProxy != "http://uppercase-host:8443" {
			t.Errorf("uppercase HttpsProxy should be preserved, got %q", config.HttpsProxy)
		}
		if config.NoProxy != "10.0.0.0/8" {
			t.Errorf("uppercase NoProxy should be preserved, got %q", config.NoProxy)
		}
	})
}
