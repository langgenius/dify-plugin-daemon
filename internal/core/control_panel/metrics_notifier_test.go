package controlpanel

import (
	"errors"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/assert"
)

func TestPluginIDFromIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "standard plugin identifier",
			identifier: "langgenius/openai:1.0.0@abc123def456789abc123def456789abcd",
			expected:   "langgenius/openai",
		},
		{
			name:       "plugin without author",
			identifier: "openai:1.0.0@abc123def456789abc123def456789abcd",
			expected:   "openai",
		},
		{
			name:       "complex plugin identifier",
			identifier: "author/my-plugin:2.1.3-beta@1234567890abcdef1234567890abcdef",
			expected:   "author/my-plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identifier, err := plugin_entities.NewPluginUniqueIdentifier(tt.identifier)
			assert.NoError(t, err)

			result := identifier.PluginID()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewMetricsNotifier(t *testing.T) {
	notifier := NewMetricsNotifier()
	assert.NotNil(t, notifier)

	assert.IsType(t, &MetricsNotifier{}, notifier, "NewMetricsNotifier should return *MetricsNotifier type")
}

func TestMetricsNotifier_OnLocalRuntimeStarting(t *testing.T) {
	notifier := NewMetricsNotifier()

	identifier, err := plugin_entities.NewPluginUniqueIdentifier("langgenius/openai:1.0.0@abc123def456789abc123def456789abcd")
	assert.NoError(t, err)

	// This should not panic
	assert.NotPanics(t, func() {
		notifier.OnLocalRuntimeStarting(identifier)
	})
}

func TestMetricsNotifier_OnLocalRuntimeStartFailed(t *testing.T) {
	notifier := NewMetricsNotifier()

	identifier, err := plugin_entities.NewPluginUniqueIdentifier("langgenius/openai:1.0.0@abc123def456789abc123def456789abcd")
	assert.NoError(t, err)

	// This should not panic
	assert.NotPanics(t, func() {
		notifier.OnLocalRuntimeStartFailed(identifier, assert.AnError)
	})
}

func TestMetricsNotifier_ScaleEvents(t *testing.T) {
	notifier := NewMetricsNotifier()

	// Create a mock runtime (we can't easily create a real one in tests)
	// but we can test that the methods don't panic
	assert.NotPanics(t, func() {
		// These methods have empty implementations currently
		// but we test they exist and don't panic
		type mockRuntime struct{}
		notifier.OnLocalRuntimeScaleUp(nil, 1)
		notifier.OnLocalRuntimeScaleDown(nil, 1)
	})
}

func TestPluginIDFromRuntime(t *testing.T) {
	t.Run("nil runtime", func(t *testing.T) {
		result := pluginIDFromRuntime(nil)
		assert.Equal(t, "unknown", result)
	})

	t.Run("mock runtime with successful identity", func(t *testing.T) {
		identifier, _ := plugin_entities.NewPluginUniqueIdentifier("test-plugin:1.0.0@abc123def456789abc123def456789abcd")

		mockRuntime := &mockIdentifiableRuntime{
			identity: identifier,
			err:      nil,
		}

		result := pluginIDFromRuntime(mockRuntime)
		assert.Equal(t, "test-plugin", result)
	})

	t.Run("mock runtime with error", func(t *testing.T) {
		mockRuntime := &mockIdentifiableRuntime{
			identity: "",
			err:      errors.New("identity error"),
		}

		result := pluginIDFromRuntime(mockRuntime)
		assert.Equal(t, "unknown", result)
	})
}

// mockIdentifiableRuntime is a mock implementation of identifiableRuntime for testing
type mockIdentifiableRuntime struct {
	identity plugin_entities.PluginUniqueIdentifier
	err      error
}

func (m *mockIdentifiableRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	return m.identity, m.err
}
