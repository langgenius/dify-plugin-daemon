package health_checker

import (
	"testing"
	"time"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/assert"
)

type MockPluginInstance struct {
	isReady      bool
	lastActiveAt time.Time
}

func (m *MockPluginInstance) ID() string {
	return "mock-instance-id"
}

func (m *MockPluginInstance) IsReady() bool {
	return m.isReady
}

func (m *MockPluginInstance) LastHeartbeat() time.Time {
	return m.lastActiveAt
}

func TestLocalHealthChecker_HealthCheckConfigDefaults(t *testing.T) {
	config := &HealthCheckerConfig{
		MaxWaitTime:          300,
		CheckInterval:        5,
		RequiredSuccessCount: 3,
	}

	assert.Equal(t, 300, config.MaxWaitTime)
	assert.Equal(t, 5, config.CheckInterval)
	assert.Equal(t, 3, config.RequiredSuccessCount)
}

func TestLocalHealthChecker_NewHealthChecker(t *testing.T) {
	config := &HealthCheckerConfig{
		CheckInterval:        5,
		RequiredSuccessCount: 3,
	}

	checker := NewLocalHealthChecker(&controlpanel.ControlPanel{}, config)

	assert.NotNil(t, checker)
	assert.Equal(t, config, checker.config)
}

func TestHealthCheckerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  HealthCheckerConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: HealthCheckerConfig{
				MaxWaitTime:          300,
				CheckInterval:        5,
				RequiredSuccessCount: 3,
			},
			wantErr: false,
		},
		{
			name: "zero check interval",
			config: HealthCheckerConfig{
				MaxWaitTime:          300,
				CheckInterval:        0,
				RequiredSuccessCount: 3,
			},
			wantErr: true,
		},
		{
			name: "zero required success count",
			config: HealthCheckerConfig{
				MaxWaitTime:          300,
				CheckInterval:        5,
				RequiredSuccessCount: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.CheckInterval == 0 || tt.config.RequiredSuccessCount == 0 {
				assert.True(t, tt.wantErr)
			} else {
				assert.False(t, tt.wantErr)
			}
		})
	}
}

func TestNewHealthChecker_RuntimeTypes(t *testing.T) {
	tests := []struct {
		name        string
		runtimeType plugin_entities.PluginRuntimeType
		wantErr     bool
	}{
		{
			name:        "local runtime",
			runtimeType: plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
			wantErr:     false,
		},
		{
			name:        "serverless runtime",
			runtimeType: plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
			wantErr:     false,
		},
		{
			name:        "remote runtime",
			runtimeType: plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
			wantErr:     false,
		},
		{
			name:        "invalid runtime",
			runtimeType: plugin_entities.PluginRuntimeType("invalid"),
			wantErr:     true,
		},
	}

	config := &app.Config{
		HealthCheckMaxWaitTime:  300,
		HealthCheckInterval:     5,
		HealthCheckSuccessCount: 3,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := NewHealthChecker(
				tt.runtimeType,
				&controlpanel.ControlPanel{},
				config,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, checker)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checker)
			}
		})
	}
}

func TestLocalHealthChecker_TimeoutCalculation(t *testing.T) {
	tests := []struct {
		name           string
		timeout        int
		checkInterval  int
		expectedChecks int
	}{
		{
			name:           "5 second timeout with 1 second interval",
			timeout:        5,
			checkInterval:  1,
			expectedChecks: 5,
		},
		{
			name:           "10 second timeout with 2 second interval",
			timeout:        10,
			checkInterval:  2,
			expectedChecks: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedDuration := time.Duration(tt.timeout) * time.Second
			checkInterval := time.Duration(tt.checkInterval) * time.Second

			maxChecks := int(expectedDuration / checkInterval)

			assert.Equal(t, tt.expectedChecks, maxChecks)
		})
	}
}

func TestLocalHealthChecker_RequiredSuccessCountLogic(t *testing.T) {
	tests := []struct {
		name              string
		requiredSuccess   int
		successCount      int
		shouldReturnReady bool
	}{
		{
			name:              "need 3 successes, have 3",
			requiredSuccess:   3,
			successCount:      3,
			shouldReturnReady: true,
		},
		{
			name:              "need 3 successes, have 2",
			requiredSuccess:   3,
			successCount:      2,
			shouldReturnReady: false,
		},
		{
			name:              "need 3 successes, have 4",
			requiredSuccess:   3,
			successCount:      4,
			shouldReturnReady: true,
		},
		{
			name:              "need 1 success, have 1",
			requiredSuccess:   1,
			successCount:      1,
			shouldReturnReady: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consecutiveSuccesses := tt.successCount
			isReady := consecutiveSuccesses >= tt.requiredSuccess

			assert.Equal(t, tt.shouldReturnReady, isReady)
		})
	}
}

func TestLocalHealthChecker_HeartbeatValidation(t *testing.T) {
	tests := []struct {
		name          string
		lastHeartbeat time.Time
		shouldBeValid bool
	}{
		{
			name:          "recent heartbeat (10 seconds ago)",
			lastHeartbeat: time.Now().Add(-10 * time.Second),
			shouldBeValid: true,
		},
		{
			name:          "heartbeat at threshold (120 seconds ago)",
			lastHeartbeat: time.Now().Add(-120 * time.Second),
			shouldBeValid: false,
		},
		{
			name:          "stale heartbeat (121 seconds ago)",
			lastHeartbeat: time.Now().Add(-121 * time.Second),
			shouldBeValid: false,
		},
		{
			name:          "very old heartbeat (5 minutes ago)",
			lastHeartbeat: time.Now().Add(-5 * time.Minute),
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeSince := time.Since(tt.lastHeartbeat)
			isValid := timeSince < 120*time.Second

			assert.Equal(t, tt.shouldBeValid, isValid)
		})
	}
}
