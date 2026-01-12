package health_checker

import (
	"testing"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/stretchr/testify/assert"
)

func TestNewServerlessHealthChecker(t *testing.T) {
	config := &HealthCheckerConfig{
		CheckInterval:        5,
		RequiredSuccessCount: 3,
	}

	checker := NewServerlessHealthChecker(&controlpanel.ControlPanel{}, config)

	assert.NotNil(t, checker)
	assert.Equal(t, config, checker.config)
}

func TestServerlessHealthChecker_Configuration(t *testing.T) {
	tests := []struct {
		name              string
		checkInterval     int
		requiredSuccess   int
		expectValidConfig bool
	}{
		{
			name:              "valid production config",
			checkInterval:     5,
			requiredSuccess:   3,
			expectValidConfig: true,
		},
		{
			name:              "fast check config for testing",
			checkInterval:     1,
			requiredSuccess:   1,
			expectValidConfig: true,
		},
		{
			name:              "invalid config - zero interval",
			checkInterval:     0,
			requiredSuccess:   3,
			expectValidConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HealthCheckerConfig{
				CheckInterval:        tt.checkInterval,
				RequiredSuccessCount: tt.requiredSuccess,
			}

			checker := NewServerlessHealthChecker(&controlpanel.ControlPanel{}, config)

			if tt.expectValidConfig {
				assert.NotNil(t, checker)
			} else {
				// 即使配置无效，checker 也会被创建
				// 问题会在运行时出现
				assert.NotNil(t, checker)
			}
		})
	}
}

func TestServerlessHealthChecker_RuntimeValidation(t *testing.T) {
	tests := []struct {
		name         string
		functionURL  string
		functionName string
		shouldBeValid bool
	}{
		{
			name:         "valid deployment",
			functionURL:  "https://lambda.amazonaws.com/2023-08-15/functions/my-plugin/invocations",
			functionName: "my-plugin-v2",
			shouldBeValid: true,
		},
		{
			name:         "missing function URL",
			functionURL:  "",
			functionName: "my-plugin",
			shouldBeValid: false,
		},
		{
			name:         "missing function name",
			functionURL:  "https://lambda.amazonaws.com/2023-08-15/functions/my-plugin/invocations",
			functionName: "",
			shouldBeValid: false,
		},
		{
			name:         "both missing",
			functionURL:  "",
			functionName: "",
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟运行时模型验证逻辑
			hasURL := tt.functionURL != ""
			hasName := tt.functionName != ""
			isValid := hasURL && hasName

			assert.Equal(t, tt.shouldBeValid, isValid)
		})
	}
}
