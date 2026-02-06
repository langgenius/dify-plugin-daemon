package plugin_manager

import (
	"testing"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func TestReadinessNonLocal(t *testing.T) {
	pm := &PluginManager{
		config: &app.Config{Platform: app.PLATFORM_SERVERLESS},
	}
	r := pm.Readiness()
	if !r.Ready {
		t.Fatalf("expected ready for non-local platform")
	}
}

func TestReadinessLocalNoSnapshot(t *testing.T) {
	pm := &PluginManager{
		config:       &app.Config{Platform: app.PLATFORM_LOCAL},
		controlPanel: controlpanel.NewControlPanel(&app.Config{Platform: app.PLATFORM_LOCAL, PluginLocalLaunchingConcurrent: 1}, nil, nil, nil),
	}
	r := pm.Readiness()
	if r.Ready {
		t.Fatalf("expected unready when snapshot is missing")
	}
}
