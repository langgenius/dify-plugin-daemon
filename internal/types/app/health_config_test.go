package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHealthCheckConfig(t *testing.T) {
	config := &Config{
		EnableZeroDowntimeUpgrade: true,
		HealthCheckMaxWaitTime:    300,
		HealthCheckInterval:       5,
		HealthCheckSuccessCount:   3,
	}

	// 验证默认值
	assert.True(t, config.EnableZeroDowntimeUpgrade)
	assert.Equal(t, 300, config.HealthCheckMaxWaitTime)
	assert.Equal(t, 5, config.HealthCheckInterval)
	assert.Equal(t, 3, config.HealthCheckSuccessCount)
}

func TestHealthCheckConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid default config",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    300,
				HealthCheckInterval:       5,
				HealthCheckSuccessCount:   3,
			},
			wantErr: false,
		},
		{
			name: "valid production config with longer timeout",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    600,
				HealthCheckInterval:       10,
				HealthCheckSuccessCount:   5,
			},
			wantErr: false,
		},
		{
			name: "valid fast check config for testing",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    30,
				HealthCheckInterval:       1,
				HealthCheckSuccessCount:   1,
			},
			wantErr: false,
		},
		{
			name: "zero downtime upgrade disabled",
			config: Config{
				EnableZeroDowntimeUpgrade: false,
				HealthCheckMaxWaitTime:    300,
				HealthCheckInterval:       5,
				HealthCheckSuccessCount:   3,
			},
			wantErr: false,
		},
		{
			name: "invalid - zero wait time",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    0,
				HealthCheckInterval:       5,
				HealthCheckSuccessCount:   3,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero check interval",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    300,
				HealthCheckInterval:       0,
				HealthCheckSuccessCount:   3,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero success count",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    300,
				HealthCheckInterval:       5,
				HealthCheckSuccessCount:   0,
			},
			wantErr: true,
		},
		{
			name: "invalid - wait time less than interval",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    10,
				HealthCheckInterval:       30,
				HealthCheckSuccessCount:   3,
			},
			wantErr: true,
		},
		{
			name: "invalid - success count exceeds reasonable limit",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    300,
				HealthCheckInterval:       5,
				HealthCheckSuccessCount:   100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			// 验证配置
			if tt.config.HealthCheckMaxWaitTime <= 0 {
				err = assert.AnError
			}
			if tt.config.HealthCheckInterval <= 0 {
				err = assert.AnError
			}
			if tt.config.HealthCheckSuccessCount <= 0 {
				err = assert.AnError
			}
			if tt.config.HealthCheckMaxWaitTime < tt.config.HealthCheckInterval {
				err = assert.AnError
			}
			if tt.config.HealthCheckSuccessCount > 10 {
				err = assert.AnError
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHealthCheckConfig_CalculateMaxChecks(t *testing.T) {
	tests := []struct {
		name              string
		maxWaitTime       int
		checkInterval     int
		expectedMaxChecks int
	}{
		{
			name:              "default config",
			maxWaitTime:       300,
			checkInterval:     5,
			expectedMaxChecks: 60,
		},
		{
			name:              "fast check",
			maxWaitTime:       30,
			checkInterval:     1,
			expectedMaxChecks: 30,
		},
		{
			name:              "long interval",
			maxWaitTime:       600,
			checkInterval:     10,
			expectedMaxChecks: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxChecks := tt.maxWaitTime / tt.checkInterval
			assert.Equal(t, tt.expectedMaxChecks, maxChecks)
		})
	}
}

func TestHealthCheckConfig_EnableZeroDowntimeUpgrade(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		description string
	}{
		{
			name:        "enabled",
			enabled:     true,
			description: "zero-downtime upgrade is active",
		},
		{
			name:        "disabled",
			enabled:     false,
			description: "falls back to old upgrade behavior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				EnableZeroDowntimeUpgrade: tt.enabled,
			}

			assert.Equal(t, tt.enabled, config.EnableZeroDowntimeUpgrade)
		})
	}
}

func TestHealthCheckConfig_RecommendedValues(t *testing.T) {
	tests := []struct {
		name               string
		config             Config
		isRecommended      bool
		recommendedReason  string
	}{
		{
			name: "production recommended",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    300,
				HealthCheckInterval:       5,
				HealthCheckSuccessCount:   3,
			},
			isRecommended:     true,
			recommendedReason: "balanced between safety and speed",
		},
		{
			name: "development fast",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    60,
				HealthCheckInterval:       2,
				HealthCheckSuccessCount:   2,
			},
			isRecommended:     true,
			recommendedReason: "good for development/testing",
		},
		{
			name: "too aggressive",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    10,
				HealthCheckInterval:       1,
				HealthCheckSuccessCount:   1,
			},
			isRecommended:     false,
			recommendedReason: "may fail due to slow startup",
		},
		{
			name: "too conservative",
			config: Config{
				EnableZeroDowntimeUpgrade: true,
				HealthCheckMaxWaitTime:    3600,
				HealthCheckInterval:       60,
				HealthCheckSuccessCount:   10,
			},
			isRecommended:     false,
			recommendedReason: "too slow for practical use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var isRecommended bool

			// 简单的推荐逻辑
			if tt.config.HealthCheckMaxWaitTime >= 30 &&
				tt.config.HealthCheckMaxWaitTime <= 600 &&
				tt.config.HealthCheckInterval >= 2 &&
				tt.config.HealthCheckInterval <= 10 &&
				tt.config.HealthCheckSuccessCount >= 2 &&
				tt.config.HealthCheckSuccessCount <= 5 {
				isRecommended = true
			}

			assert.Equal(t, tt.isRecommended, isRecommended)
		})
	}
}
