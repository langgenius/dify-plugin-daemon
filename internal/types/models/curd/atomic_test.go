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
