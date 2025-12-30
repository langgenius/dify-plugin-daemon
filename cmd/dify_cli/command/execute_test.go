package command

import (
	"reflect"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestParseToolArgs(t *testing.T) {
	tool := &plugin_entities.ToolDeclaration{
		Parameters: []plugin_entities.ToolParameter{
			{Name: "query", Type: plugin_entities.TOOL_PARAMETER_TYPE_STRING},
			{Name: "count", Type: plugin_entities.TOOL_PARAMETER_TYPE_NUMBER},
			{Name: "enabled", Type: plugin_entities.TOOL_PARAMETER_TYPE_BOOLEAN},
		},
	}

	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "string parameter",
			args: []string{"--query", "hello world"},
			want: map[string]any{"query": "hello world"},
		},
		{
			name: "number parameter",
			args: []string{"--count", "42"},
			want: map[string]any{"count": float64(42)},
		},
		{
			name: "boolean true",
			args: []string{"--enabled", "true"},
			want: map[string]any{"enabled": true},
		},
		{
			name: "boolean false",
			args: []string{"--enabled", "false"},
			want: map[string]any{"enabled": false},
		},
		{
			name: "multiple parameters",
			args: []string{"--query", "test", "--count", "10", "--enabled", "1"},
			want: map[string]any{"query": "test", "count": float64(10), "enabled": true},
		},
		{
			name: "unknown parameter treated as string",
			args: []string{"--unknown", "value"},
			want: map[string]any{"unknown": "value"},
		},
		{
			name: "empty args",
			args: []string{},
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseToolArgs(tool, tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseToolArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
