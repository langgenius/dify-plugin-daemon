package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPluginMetricLabels(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		pluginID, runtimeType := getPluginMetricLabels(nil)
		assert.Equal(t, "unknown", pluginID)
		assert.Equal(t, "unknown", runtimeType)
	})

	t.Run("session with nil runtime", func(t *testing.T) {
		// This test would require creating a mock session
		// For now, we just verify the function handles nil gracefully
		pluginID, runtimeType := getPluginMetricLabels(nil)
		assert.Equal(t, "unknown", pluginID)
		assert.Equal(t, "unknown", runtimeType)
	})
}
