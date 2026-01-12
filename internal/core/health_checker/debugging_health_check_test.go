package health_checker

import (
	"testing"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/stretchr/testify/assert"
)

func TestNewDebuggingHealthChecker(t *testing.T) {
	config := &HealthCheckerConfig{
		CheckInterval:        5,
		RequiredSuccessCount: 3,
	}

	checker := NewDebuggingHealthChecker(&controlpanel.ControlPanel{}, config)

	assert.NotNil(t, checker)
	assert.Equal(t, config, checker.config)
}

func TestDebuggingHealthChecker_Configuration(t *testing.T) {
	tests := []struct {
		name              string
		checkInterval     int
		requiredSuccess   int
	}{
		{
			name:            "standard config",
			checkInterval:   5,
			requiredSuccess: 3,
		},
		{
			name:            "fast check for development",
			checkInterval:   1,
			requiredSuccess: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HealthCheckerConfig{
				CheckInterval:        tt.checkInterval,
				RequiredSuccessCount: tt.requiredSuccess,
			}

			checker := NewDebuggingHealthChecker(&controlpanel.ControlPanel{}, config)

			assert.NotNil(t, checker)
			assert.Equal(t, tt.checkInterval, checker.config.CheckInterval)
			assert.Equal(t, tt.requiredSuccess, checker.config.RequiredSuccessCount)
		})
	}
}
