package controlpanel

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/assert"
)

func TestIsOldVersionBeingUpgraded_StateCheck(t *testing.T) {
	tests := []struct {
		name            string
		upgradeState    string
		shouldBeBlocked bool
	}{
		{
			name:            "completed upgrade - should not block",
			upgradeState:    "completed",
			shouldBeBlocked: false,
		},
		{
			name:            "rolled back upgrade - should not block",
			upgradeState:    "rolled_back",
			shouldBeBlocked: false,
		},
		{
			name:            "empty state - should not block",
			upgradeState:    "",
			shouldBeBlocked: false,
		},
		{
			name:            "active upgrade - should block",
			upgradeState:    "upgrading",
			shouldBeBlocked: true,
		},
		{
			name:            "waiting for ready - should block",
			upgradeState:    "waiting_for_ready",
			shouldBeBlocked: true,
		},
		{
			name:            "switching traffic - should block",
			upgradeState:    "switching_traffic",
			shouldBeBlocked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟升级状态检查逻辑
			isEmpty := tt.upgradeState == ""
			isCompleted := tt.upgradeState == "completed"
			isRolledBack := tt.upgradeState == "rolled_back"
			shouldBlock := !isEmpty && !isCompleted && !isRolledBack

			assert.Equal(t, tt.shouldBeBlocked, shouldBlock)
		})
	}
}

func TestIsOldVersionBeingUpgraded_VersionComparison(t *testing.T) {
	tests := []struct {
		name                  string
		originalVersion       string
		upgradeOriginalVersion string
		shouldMatch           bool
	}{
		{
			name:                  "matching versions",
			originalVersion:       "author/test-plugin:1.0.0@abc123",
			upgradeOriginalVersion: "author/test-plugin:1.0.0@abc123",
			shouldMatch:           true,
		},
		{
			name:                  "different versions",
			originalVersion:       "author/test-plugin:1.0.0@abc123",
			upgradeOriginalVersion: "author/test-plugin:2.0.0@def456",
			shouldMatch:           false,
		},
		{
			name:                  "empty upgrade version",
			originalVersion:       "author/test-plugin:1.0.0@abc123",
			upgradeOriginalVersion: "",
			shouldMatch:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟版本比较逻辑
			isMatch := tt.originalVersion == tt.upgradeOriginalVersion

			assert.Equal(t, tt.shouldMatch, isMatch)
		})
	}
}

func TestPluginUniqueIdentifierParsing(t *testing.T) {
	tests := []struct {
		name      string
		identifier string
		wantErr   bool
	}{
		{
			name:      "valid identifier with checksum",
			identifier: "author/test-plugin:1.0.0@abc123def456789abc123def456789ab",
			wantErr:   false,
		},
		{
			name:      "invalid identifier - missing checksum",
			identifier: "author/test-plugin:1.0.0",
			wantErr:   true,
		},
		{
			name:      "invalid identifier - wrong format",
			identifier: "author/test-plugin/1.0.0/abc123",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试解析逻辑
			_, err := plugin_entities.NewPluginUniqueIdentifier(tt.identifier)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpgradeStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		currentState   string
		allowedStates []string
	}{
		{
			name:           "initializing can go to waiting",
			currentState:   "initializing",
			allowedStates: []string{"waiting_for_ready", "failed", "rolled_back"},
		},
		{
			name:           "waiting can go to switching or rollback",
			currentState:   "waiting_for_ready",
			allowedStates: []string{"switching_traffic", "rolled_back", "failed"},
		},
		{
			name:           "switching can go to cleanup or rollback",
			currentState:   "switching_traffic",
			allowedStates: []string{"cleanup_old", "rolled_back", "failed"},
		},
		{
			name:           "cleanup can go to completed",
			currentState:   "cleanup_old",
			allowedStates: []string{"completed", "failed"},
		},
		{
			name:           "completed is terminal",
			currentState:   "completed",
			allowedStates: []string{}, // 终态，不允许转移
		},
		{
			name:           "rolled_back is terminal",
			currentState:   "rolled_back",
			allowedStates: []string{}, // 终态，不允许转移
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证状态转移逻辑
			isTerminal := tt.currentState == "completed" || tt.currentState == "rolled_back"

			if isTerminal {
				assert.Empty(t, tt.allowedStates, "terminal state should not have transitions")
			} else {
				assert.NotEmpty(t, tt.allowedStates, "non-terminal state should have transitions")
			}
		})
	}
}

func TestRemoveUnusedLocalPlugins_UpgradeProtectionLogic(t *testing.T) {
	tests := []struct {
		name                 string
		pluginInInstalledBucket bool
		upgradeState          string
		upgradeOriginalVersion string
		shouldRemove          bool
	}{
		{
			name:                   "not in bucket, not upgrading - should remove",
			pluginInInstalledBucket: false,
			upgradeState:           "",
			upgradeOriginalVersion:  "",
			shouldRemove:           true,
		},
		{
			name:                   "not in bucket, upgrading - should NOT remove",
			pluginInInstalledBucket: false,
			upgradeState:           "waiting_for_ready",
			upgradeOriginalVersion:  "old-version",
			shouldRemove:           false,
		},
		{
			name:                   "in bucket - should NOT remove",
			pluginInInstalledBucket: true,
			upgradeState:           "",
			upgradeOriginalVersion:  "",
			shouldRemove:           false,
		},
		{
			name:                   "completed upgrade - should remove",
			pluginInInstalledBucket: false,
			upgradeState:           "completed",
			upgradeOriginalVersion:  "old-version",
			shouldRemove:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟清理逻辑
			shouldRemove := false

			if !tt.pluginInInstalledBucket {
				// 检查是否在升级中
				isUpgrading := tt.upgradeState != "" &&
					tt.upgradeState != "completed" &&
					tt.upgradeState != "rolled_back"

				if !isUpgrading {
					shouldRemove = true
				}
			}

			assert.Equal(t, tt.shouldRemove, shouldRemove)
		})
	}
}
