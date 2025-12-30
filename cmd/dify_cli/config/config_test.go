package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/cmd/dify_cli/types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestLoadEnvFile(t *testing.T) {
	content := `INNER_API_URL=http://localhost:5001
INNER_API_KEY=test-key-123
# comment line
IGNORED_VAR=should_be_ignored
`
	tmpFile := filepath.Join(t.TempDir(), "test.env")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadEnvFile failed: %v", err)
	}

	if cfg.InnerAPIURL != "http://localhost:5001" {
		t.Errorf("InnerAPIURL = %q, want %q", cfg.InnerAPIURL, "http://localhost:5001")
	}
	if cfg.InnerAPIKey != "test-key-123" {
		t.Errorf("InnerAPIKey = %q, want %q", cfg.InnerAPIKey, "test-key-123")
	}
}

func TestLoadEnvFileWithQuotes(t *testing.T) {
	content := `INNER_API_URL="http://localhost:5001"
INNER_API_KEY='test-key-123'
`
	tmpFile := filepath.Join(t.TempDir(), "test.env")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadEnvFile failed: %v", err)
	}

	if cfg.InnerAPIURL != "http://localhost:5001" {
		t.Errorf("InnerAPIURL = %q, want %q", cfg.InnerAPIURL, "http://localhost:5001")
	}
	if cfg.InnerAPIKey != "test-key-123" {
		t.Errorf("InnerAPIKey = %q, want %q", cfg.InnerAPIKey, "test-key-123")
	}
}

func TestLoadSchemaFile(t *testing.T) {
	content := `[
  {
    "identity": {
      "author": "test",
      "name": "test_provider",
      "description": {"en_US": "Test provider"},
      "icon": "icon.svg",
      "label": {"en_US": "Test"}
    },
    "credentials_schema": [],
    "tools": [
      {
        "identity": {
          "author": "test",
          "name": "test_tool",
          "label": {"en_US": "Test Tool"}
        },
        "description": {
          "human": {"en_US": "A test tool"},
          "llm": "A test tool for testing"
        },
        "parameters": [
          {
            "name": "query",
            "label": {"en_US": "Query"},
            "human_description": {"en_US": "The query"},
            "type": "string",
            "form": "llm",
            "required": true
          }
        ]
      }
    ]
  }
]`
	tmpFile := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	providers, err := LoadSchemaFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadSchemaFile failed: %v", err)
	}

	if len(providers) != 1 {
		t.Fatalf("len(providers) = %d, want 1", len(providers))
	}

	if providers[0].Identity.Name != "test_provider" {
		t.Errorf("provider name = %q, want %q", providers[0].Identity.Name, "test_provider")
	}

	if len(providers[0].Tools) != 1 {
		t.Fatalf("len(tools) = %d, want 1", len(providers[0].Tools))
	}

	if providers[0].Tools[0].Identity.Name != "test_tool" {
		t.Errorf("tool name = %q, want %q", providers[0].Tools[0].Identity.Name, "test_tool")
	}
}

func TestFindTool(t *testing.T) {
	cfg := &types.DifyConfig{
		Providers: []plugin_entities.ToolProviderDeclaration{
			{
				Identity: plugin_entities.ToolProviderIdentity{Name: "provider1"},
				Tools: []plugin_entities.ToolDeclaration{
					{Identity: plugin_entities.ToolIdentity{Name: "tool_a"}},
					{Identity: plugin_entities.ToolIdentity{Name: "tool_b"}},
				},
			},
			{
				Identity: plugin_entities.ToolProviderIdentity{Name: "provider2"},
				Tools: []plugin_entities.ToolDeclaration{
					{Identity: plugin_entities.ToolIdentity{Name: "tool_c"}},
				},
			},
		},
	}

	tests := []struct {
		name         string
		wantProvider string
		wantTool     string
		wantFound    bool
	}{
		{"tool_a", "provider1", "tool_a", true},
		{"tool_b", "provider1", "tool_b", true},
		{"tool_c", "provider2", "tool_c", true},
		{"not_exist", "", "", false},
	}

	for _, tt := range tests {
		provider, tool := FindTool(cfg, tt.name)
		if tt.wantFound {
			if provider == nil || tool == nil {
				t.Errorf("FindTool(%q) = nil, want found", tt.name)
				continue
			}
			if provider.Identity.Name != tt.wantProvider {
				t.Errorf("FindTool(%q) provider = %q, want %q", tt.name, provider.Identity.Name, tt.wantProvider)
			}
			if tool.Identity.Name != tt.wantTool {
				t.Errorf("FindTool(%q) tool = %q, want %q", tt.name, tool.Identity.Name, tt.wantTool)
			}
		} else {
			if provider != nil || tool != nil {
				t.Errorf("FindTool(%q) = found, want nil", tt.name)
			}
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &types.DifyConfig{
		Env: types.EnvConfig{
			InnerAPIURL: "http://test:5001",
			InnerAPIKey: "test-key",
		},
		Providers: []plugin_entities.ToolProviderDeclaration{
			{
				Identity: plugin_entities.ToolProviderIdentity{Name: "test_provider"},
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Env.InnerAPIURL != cfg.Env.InnerAPIURL {
		t.Errorf("InnerAPIURL = %q, want %q", loaded.Env.InnerAPIURL, cfg.Env.InnerAPIURL)
	}
	if loaded.Env.InnerAPIKey != cfg.Env.InnerAPIKey {
		t.Errorf("InnerAPIKey = %q, want %q", loaded.Env.InnerAPIKey, cfg.Env.InnerAPIKey)
	}
}
