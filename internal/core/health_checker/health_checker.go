package health_checker

import (
	"fmt"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type HealthCheckerConfig struct {
	MaxWaitTime          int `json:"max_wait_time" default:"300"`
	CheckInterval        int `json:"check_interval" default:"5"`
	RequiredSuccessCount int `json:"required_success_count" default:"3"`
}

type PluginHealthChecker interface {
	IsReady(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) (bool, error)
	WaitForReady(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier, timeout int) (bool, error)
}

func NewHealthChecker(
	runtimeType plugin_entities.PluginRuntimeType,
	controlPanel *controlpanel.ControlPanel,
	config *app.Config,
) (PluginHealthChecker, error) {
	cfg := &HealthCheckerConfig{
		MaxWaitTime:          config.HealthCheckMaxWaitTime,
		CheckInterval:        config.HealthCheckInterval,
		RequiredSuccessCount: config.HealthCheckSuccessCount,
	}

	switch runtimeType {
	case plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL:
		return NewLocalHealthChecker(controlPanel, cfg), nil
	case plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS:
		return NewServerlessHealthChecker(controlPanel, cfg), nil
	case plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE:
		return NewDebuggingHealthChecker(controlPanel, cfg), nil
	default:
		return nil, fmt.Errorf("unsupported runtime type: %s", runtimeType)
	}
}
