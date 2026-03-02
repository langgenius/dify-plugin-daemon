package plugin_manager

import (
	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

type ReadinessReport struct {
	Ready   bool
	Reason  string
	Plugins *controlpanel.LocalReadinessSnapshot
}

func (p *PluginManager) Readiness() ReadinessReport {
	if p == nil || p.config == nil {
		return ReadinessReport{Ready: false, Reason: "manager_not_initialized"}
	}

	if p.config.Platform != app.PLATFORM_LOCAL {
		return ReadinessReport{Ready: true, Reason: "non_local_platform"}
	}

	snapshot, ok := p.controlPanel.LocalReadiness()
	if !ok {
		return ReadinessReport{Ready: false, Reason: "plugin_monitor_not_ready"}
	}

	if snapshot.Ready {
		return ReadinessReport{Ready: true, Reason: "plugins_ready", Plugins: &snapshot}
	}
	if len(snapshot.Failed) > 0 {
		return ReadinessReport{Ready: false, Reason: "plugins_failed", Plugins: &snapshot}
	}
	return ReadinessReport{Ready: false, Reason: "plugins_starting", Plugins: &snapshot}
}
