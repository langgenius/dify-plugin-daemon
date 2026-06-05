package service

import (
	"encoding/json"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestUpgradePlugin(t *testing.T) {
	originalIdentifier, err := plugin_entities.NewPluginUniqueIdentifier("author/test-plugin:1.0.0@abcdef1234567890abcdef1234567890ab")
	if err != nil {
		t.Fatalf("failed to create original plugin unique identifier: %v", err)
	}

	newIdentifier, err := plugin_entities.NewPluginUniqueIdentifier("author/test-plugin:2.0.0@1234567890abcdef1234567890abcdef12")
	if err != nil {
		t.Fatalf("failed to create new plugin unique identifier: %v", err)
	}

	config := &app.Config{
		PluginInstallTimeout: 15,
	}

	tests := []struct {
		name                             string
		tenantId                         string
		source                           string
		meta                             map[string]any
		originalPluginUniqueIdentifier   plugin_entities.PluginUniqueIdentifier
		newPluginUniqueIdentifier        plugin_entities.PluginUniqueIdentifier
		wantSuccess                      bool
		wantAllInstalled                 bool
		wantTaskIDEmpty                  bool
	}{
		{
			name:                           "same plugin identifiers",
			tenantId:                       "tenant-123",
			source:                         "test",
			meta:                           map[string]any{},
			originalPluginUniqueIdentifier:  originalIdentifier,
			newPluginUniqueIdentifier:       originalIdentifier,
			wantSuccess:                    false,
		},
		{
			name:                           "different plugin identifiers",
			tenantId:                       "tenant-123",
			source:                         "test",
			meta:                           map[string]any{},
			originalPluginUniqueIdentifier:  originalIdentifier,
			newPluginUniqueIdentifier:       newIdentifier,
			wantSuccess:                    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := UpgradePlugin(
				config,
				tt.tenantId,
				tt.source,
				tt.meta,
				tt.originalPluginUniqueIdentifier,
				tt.newPluginUniqueIdentifier,
			)

			isSuccess := response.Code == 0
			if isSuccess != tt.wantSuccess {
				t.Errorf("UpgradePlugin() success = %v, want %v", isSuccess, tt.wantSuccess)
			}

			if isSuccess {
				var result InstallPluginResponse
				dataBytes, err := json.Marshal(response.Data)
				if err == nil {
					_ = json.Unmarshal(dataBytes, &result)
					if tt.wantAllInstalled && !result.AllInstalled {
						t.Errorf("UpgradePlugin() AllInstalled = %v, want %v", result.AllInstalled, tt.wantAllInstalled)
					}
					if tt.wantTaskIDEmpty && result.TaskID != "" {
						t.Errorf("UpgradePlugin() TaskID = %v, want empty", result.TaskID)
					}
				}
			}
		})
	}
}
