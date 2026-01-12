package health_checker

import (
	"fmt"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type DebuggingHealthChecker struct {
	controlPanel *controlpanel.ControlPanel
	config       *HealthCheckerConfig
}

func NewDebuggingHealthChecker(
	controlPanel *controlpanel.ControlPanel,
	config *HealthCheckerConfig,
) *DebuggingHealthChecker {
	return &DebuggingHealthChecker{
		controlPanel: controlPanel,
		config:       config,
	}
}

func (h *DebuggingHealthChecker) IsReady(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	_, exists := h.controlPanel.GetDebuggingPluginRuntime(pluginUniqueIdentifier)
	if !exists {
		return false, fmt.Errorf("debug runtime not found for plugin: %s", pluginUniqueIdentifier.String())
	}

	return true, nil
}

func (h *DebuggingHealthChecker) WaitForReady(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	timeout int,
) (bool, error) {
	ready, err := h.IsReady(pluginUniqueIdentifier)
	if err != nil {
		return false, err
	}
	return ready, nil
}
