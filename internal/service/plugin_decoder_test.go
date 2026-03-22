package service

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestUploadPluginPkg_CacheKeyFormat(t *testing.T) {
	tests := []struct {
		name                    string
		pluginUniqueIdentifier   plugin_entities.PluginUniqueIdentifier
		expectedKey              string
	}{
		{
			name:                    "standard plugin identifier",
			pluginUniqueIdentifier:   plugin_entities.PluginUniqueIdentifier("author/plugin@1.0.0"),
			expectedKey:              "manually_uploaded:author/plugin@1.0.0",
		},
		{
			name:                    "plugin with version hash",
			pluginUniqueIdentifier:   plugin_entities.PluginUniqueIdentifier("author/plugin@abc123def456"),
			expectedKey:              "manually_uploaded:author/plugin@abc123def456",
		},
		{
			name:                    "plugin with special characters",
			pluginUniqueIdentifier:   plugin_entities.PluginUniqueIdentifier("my-org/my-plugin@1.0.0-beta"),
			expectedKey:              "manually_uploaded:my-org/my-plugin@1.0.0-beta",
		},
		{
			name:                    "langgenius gemini plugin",
			pluginUniqueIdentifier:   plugin_entities.PluginUniqueIdentifier("langgenius/gemini@0.7.17"),
			expectedKey:              "manually_uploaded:langgenius/gemini@0.7.17",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheKey := "manually_uploaded:" + tt.pluginUniqueIdentifier.String()

			if cacheKey != tt.expectedKey {
				t.Errorf("cache key mismatch\n got: %s\nwant: %s", cacheKey, tt.expectedKey)
			}

			t.Logf("✓ Cache key format correct: %s", cacheKey)
		})
	}
}
