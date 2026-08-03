package generator

import (
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers/definitions"
)

func TestPluginDaemonOutputPath(t *testing.T) {
	got := pluginDaemonOutputPath(access_types.PLUGIN_ACCESS_TYPE_MODEL)
	want := filepath.Join("internal", "core", "io_tunnel", "model.gen.go")

	if got != want {
		t.Fatalf("pluginDaemonOutputPath() = %q, want %q", got, want)
	}
}

func TestPluginDaemonTemplateRoutesOnlyConfiguredCustomHandlers(t *testing.T) {
	dispatchers := []*definitions.PluginDispatcher{
		{
			Name:                "CustomHandler",
			RequestTypeString:   "requests.CustomRequest",
			ResponseTypeString:  "responses.CustomResponse",
			BufferSize:          1,
			CustomTunnelHandler: "customTunnelHandler",
		},
		{
			Name:               "GenericHandler",
			RequestTypeString:  "requests.GenericRequest",
			ResponseTypeString: "responses.GenericResponse",
			BufferSize:         1,
		},
	}
	tmpl := template.Must(template.New("pluginDaemon").Parse(pluginDaemonTemplate))
	var output strings.Builder

	err := tmpl.Execute(&output, struct {
		AccessType  access_types.PluginAccessType
		Dispatchers []*definitions.PluginDispatcher
	}{
		AccessType:  access_types.PLUGIN_ACCESS_TYPE_MODEL,
		Dispatchers: dispatchers,
	})
	if err != nil {
		t.Fatalf("execute plugin daemon template: %v", err)
	}

	generated := output.String()
	if !strings.Contains(generated, "return customTunnelHandler(") {
		t.Fatalf("custom handler was not generated:\n%s", generated)
	}
	if count := strings.Count(generated, "return GenericInvokePlugin["); count != 1 {
		t.Fatalf("GenericInvokePlugin call count = %d, want 1:\n%s", count, generated)
	}
}
