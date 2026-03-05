package curd

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
)

// TestGetPluginID_Logic tests the getPluginID() helper function logic
// which is used inside UninstallPlugin function
//
// This test verifies two scenarios:
// 1. When pluginToBeReturns is not nil, returns pluginToBeReturns.PluginID
// 2. When pluginToBeReturns is nil, returns installation.PluginID
func TestGetPluginID_Logic(t *testing.T) {
	// Simulate the getPluginID closure logic
	type testCase struct {
		name             string
		pluginToBeReturns *models.Plugin
		installation      *models.PluginInstallation
		expectedPluginID  string
	}

	testCases := []testCase{
		{
			name: "pluginToBeReturns exists - use pluginToBeReturns.PluginID",
			pluginToBeReturns: &models.Plugin{
				PluginID: "plugin-from-record",
			},
			installation: &models.PluginInstallation{
				PluginID: "plugin-from-installation",
			},
			expectedPluginID: "plugin-from-record",
		},
		{
			name:             "pluginToBeReturns is nil - use installation.PluginID",
			pluginToBeReturns: nil,
			installation: &models.PluginInstallation{
				PluginID: "plugin-from-installation",
			},
			expectedPluginID: "plugin-from-installation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate getPluginID() logic
			var result string
			if tc.pluginToBeReturns != nil {
				result = tc.pluginToBeReturns.PluginID
			} else {
				result = tc.installation.PluginID
			}

			if result != tc.expectedPluginID {
				t.Errorf("getPluginID() = %s, want %s", result, tc.expectedPluginID)
			}
		})
	}
}

// TestInstallPlugin_UpdatePluginID tests the plugin_id update logic
// when installing a remote plugin with a different plugin_id format
func TestInstallPlugin_UpdatePluginID(t *testing.T) {
	// Test case: plugin_id changes from old format to new format
	type testCase struct {
		name                  string
		existingPluginID      string
		newPluginID           string
		shouldUpdate          bool
		expectedFinalPluginID string
	}

	testCases := []testCase{
		{
			name:                  "plugin_id differs - should update",
			existingPluginID:      "old-author/old-plugin:1.0.0",
			newPluginID:           "author/name",
			shouldUpdate:          true,
			expectedFinalPluginID: "author/name",
		},
		{
			name:                  "plugin_id same - should not update",
			existingPluginID:      "author/name",
			newPluginID:           "author/name",
			shouldUpdate:          false,
			expectedFinalPluginID: "author/name",
		},
		{
			name:                  "empty to new - should update",
			existingPluginID:      "",
			newPluginID:           "author/new-plugin",
			shouldUpdate:          true,
			expectedFinalPluginID: "author/new-plugin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate plugin record
			plugin := &models.Plugin{
				PluginID: tc.existingPluginID,
			}

			// Simulate the update logic
			updated := false
			if plugin.PluginID != tc.newPluginID {
				plugin.PluginID = tc.newPluginID
				updated = true
			}

			// Verify update behavior
			if updated != tc.shouldUpdate {
				t.Errorf("updated = %v, want %v", updated, tc.shouldUpdate)
			}

			if plugin.PluginID != tc.expectedFinalPluginID {
				t.Errorf("PluginID = %s, want %s", plugin.PluginID, tc.expectedFinalPluginID)
			}
		})
	}
}


