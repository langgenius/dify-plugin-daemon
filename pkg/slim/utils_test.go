package slim

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 0, "..."},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q; want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestPluginWorkingPath(t *testing.T) {
	tests := []struct {
		folder   string
		pluginID string
		wantEnd  string
	}{
		{"/plugins", "author/plugin:1.0.0", "author/plugin-1.0.0"},
		{"/plugins", "simple", "simple"},
		{"/plugins", "a:b:c", "a-b-c"},
	}
	for _, tt := range tests {
		got := pluginWorkingPath(tt.folder, tt.pluginID)
		if got != tt.folder+"/"+tt.wantEnd {
			t.Errorf("pluginWorkingPath(%q, %q) = %q; want suffix %q",
				tt.folder, tt.pluginID, got, tt.wantEnd)
		}
	}
}

func TestEnv(t *testing.T) {
	t.Setenv("SLIM_TEST_VAR", "value")
	if got := env("SLIM_TEST_VAR", "default"); got != "value" {
		t.Errorf("env() = %q; want %q", got, "value")
	}
	if got := env("SLIM_TEST_UNSET_VAR", "default"); got != "default" {
		t.Errorf("env() = %q; want %q", got, "default")
	}
}

func TestEnvInt(t *testing.T) {
	t.Setenv("SLIM_TEST_INT", "42")
	if got := envInt("SLIM_TEST_INT", 0); got != 42 {
		t.Errorf("envInt() = %d; want 42", got)
	}

	t.Setenv("SLIM_TEST_INT_BAD", "abc")
	if got := envInt("SLIM_TEST_INT_BAD", 99); got != 99 {
		t.Errorf("envInt() = %d; want 99 (default on parse error)", got)
	}

	if got := envInt("SLIM_TEST_INT_UNSET", 10); got != 10 {
		t.Errorf("envInt() = %d; want 10", got)
	}
}
