package health_checker

import (
	"fmt"
	"time"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type LocalHealthChecker struct {
	controlPanel *controlpanel.ControlPanel
	config       *HealthCheckerConfig
}

func NewLocalHealthChecker(
	controlPanel *controlpanel.ControlPanel,
	config *HealthCheckerConfig,
) *LocalHealthChecker {
	return &LocalHealthChecker{
		controlPanel: controlPanel,
		config:       config,
	}
}

func (h *LocalHealthChecker) IsReady(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	runtime, exists := h.controlPanel.GetLocalPluginRuntime(pluginUniqueIdentifier)
	if !exists {
		return false, fmt.Errorf("runtime not found for plugin: %s", pluginUniqueIdentifier.String())
	}

	instances := runtime.Instances()
	if len(instances) == 0 {
		return false, fmt.Errorf("no instances running for plugin: %s", pluginUniqueIdentifier.String())
	}

	hasReadyInstance := false
	for _, instance := range instances {
		if instance.IsReady() {
			lastHeartbeat := instance.LastHeartbeat()
			if time.Since(lastHeartbeat) < 120*time.Second {
				hasReadyInstance = true
				break
			}
		}
	}

	if !hasReadyInstance {
		return false, fmt.Errorf("no ready instances for plugin: %s", pluginUniqueIdentifier.String())
	}

	if runtime.Status() != "running" {
		return false, fmt.Errorf("runtime not running, status: %s", runtime.Status())
	}

	return true, nil
}

func (h *LocalHealthChecker) WaitForReady(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	timeout int,
) (bool, error) {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	checkInterval := time.Duration(h.config.CheckInterval) * time.Second
	consecutiveSuccesses := 0

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		if time.Now().After(deadline) {
			return false, fmt.Errorf("timeout waiting for plugin to be ready: %s", pluginUniqueIdentifier.String())
		}

		ready, err := h.IsReady(pluginUniqueIdentifier)
		if err != nil {
			consecutiveSuccesses = 0
			log.Debug("Health check failed for plugin %s: %v", pluginUniqueIdentifier.String(), err)
			<-ticker.C
			continue
		}

		if ready {
			consecutiveSuccesses++
			if consecutiveSuccesses >= h.config.RequiredSuccessCount {
				log.Info("Plugin %s is ready after %d consecutive health checks", pluginUniqueIdentifier.String(), consecutiveSuccesses)
				return true, nil
			}
		} else {
			consecutiveSuccesses = 0
		}

		<-ticker.C
	}
}
