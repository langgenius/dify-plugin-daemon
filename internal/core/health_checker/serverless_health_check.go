package health_checker

import (
	"fmt"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type ServerlessHealthChecker struct {
	controlPanel *controlpanel.ControlPanel
	config       *HealthCheckerConfig
}

func NewServerlessHealthChecker(
	controlPanel *controlpanel.ControlPanel,
	config *HealthCheckerConfig,
) *ServerlessHealthChecker {
	return &ServerlessHealthChecker{
		controlPanel: controlPanel,
		config:       config,
	}
}

func (h *ServerlessHealthChecker) IsReady(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	runtimeModel, err := db.GetOne[models.ServerlessRuntime](
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
	)
	if err != nil {
		return false, fmt.Errorf("serverless runtime not found: %w", err)
	}

	if runtimeModel.FunctionURL == "" || runtimeModel.FunctionName == "" {
		return false, fmt.Errorf("serverless function not properly deployed")
	}

	return true, nil
}

func (h *ServerlessHealthChecker) WaitForReady(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	_ int,
) (bool, error) {
	ready, err := h.IsReady(pluginUniqueIdentifier)
	if err != nil {
		return false, err
	}
	return ready, nil
}
