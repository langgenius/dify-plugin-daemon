package slim

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFillDefaults_LocalRequiresFolder(t *testing.T) {
	cfg := &SlimConfig{Mode: ModeLocal}
	err := fillDefaults(cfg)
	if err == nil {
		t.Fatal("fillDefaults() should fail when local.folder is empty")
	}
	se, ok := err.(*SlimError)
	if !ok {
		t.Fatalf("expected *SlimError, got %T", err)
	}
	if se.Code != ErrConfigInvalid {
		t.Fatalf("Code = %q; want %q", se.Code, ErrConfigInvalid)
	}
}

func TestFillDefaults_LocalDefaults(t *testing.T) {
	cfg := &SlimConfig{
		Mode: ModeLocal,
		Local: LocalConfig{
			Folder: "/tmp/test-plugins",
		},
	}
	if err := fillDefaults(cfg); err != nil {
		t.Fatalf("fillDefaults() error: %v", err)
	}
	if cfg.Local.PythonPath != "python3" {
		t.Errorf("PythonPath = %q; want %q", cfg.Local.PythonPath, "python3")
	}
	if cfg.Local.PythonEnvInitTimeout != 120 {
		t.Errorf("PythonEnvInitTimeout = %d; want 120", cfg.Local.PythonEnvInitTimeout)
	}
	if cfg.Local.MaxExecutionTimeout != 600 {
		t.Errorf("MaxExecutionTimeout = %d; want 600", cfg.Local.MaxExecutionTimeout)
	}
	if cfg.Local.MarketplaceURL != "https://marketplace.dify.ai" {
		t.Errorf("MarketplaceURL = %q; want %q", cfg.Local.MarketplaceURL, "https://marketplace.dify.ai")
	}
}

func TestFillDefaults_LocalPreservesExplicitValues(t *testing.T) {
	cfg := &SlimConfig{
		Mode: ModeLocal,
		Local: LocalConfig{
			Folder:               "/tmp/test-plugins",
			PythonPath:           "/usr/bin/python3.12",
			PythonEnvInitTimeout: 60,
			MaxExecutionTimeout:  300,
			MarketplaceURL:       "https://custom.marketplace.example.com",
		},
	}
	if err := fillDefaults(cfg); err != nil {
		t.Fatalf("fillDefaults() error: %v", err)
	}
	if cfg.Local.PythonPath != "/usr/bin/python3.12" {
		t.Errorf("PythonPath = %q; want %q", cfg.Local.PythonPath, "/usr/bin/python3.12")
	}
	if cfg.Local.PythonEnvInitTimeout != 60 {
		t.Errorf("PythonEnvInitTimeout = %d; want 60", cfg.Local.PythonEnvInitTimeout)
	}
	if cfg.Local.MaxExecutionTimeout != 300 {
		t.Errorf("MaxExecutionTimeout = %d; want 300", cfg.Local.MaxExecutionTimeout)
	}
	if cfg.Local.MarketplaceURL != "https://custom.marketplace.example.com" {
		t.Errorf("MarketplaceURL = %q; want custom", cfg.Local.MarketplaceURL)
	}
}

func TestFillDefaults_RemoteRequiresDaemonAddr(t *testing.T) {
	cfg := &SlimConfig{
		Mode: ModeRemote,
		Remote: RemoteConfig{
			DaemonKey: "secret",
		},
	}
	err := fillDefaults(cfg)
	if err == nil {
		t.Fatal("fillDefaults() should fail when remote.daemon_addr is empty")
	}
}

func TestFillDefaults_RemoteRequiresDaemonKey(t *testing.T) {
	cfg := &SlimConfig{
		Mode: ModeRemote,
		Remote: RemoteConfig{
			DaemonAddr: "http://localhost:5003",
		},
	}
	err := fillDefaults(cfg)
	if err == nil {
		t.Fatal("fillDefaults() should fail when remote.daemon_key is empty")
	}
}

func TestFillDefaults_EmptyModeDefaultsToRemote(t *testing.T) {
	cfg := &SlimConfig{
		Remote: RemoteConfig{
			DaemonAddr: "http://localhost:5003",
			DaemonKey:  "secret",
		},
	}
	if err := fillDefaults(cfg); err != nil {
		t.Fatalf("fillDefaults() error: %v", err)
	}
	if cfg.Mode != ModeRemote {
		t.Fatalf("Mode = %q; want %q", cfg.Mode, ModeRemote)
	}
}

func TestNewInvokeContext_Valid(t *testing.T) {
	argsJSON := `{"tenant_id":"t1","user_id":"u1","data":{"key":"val"}}`
	ctx, err := NewInvokeContext("author/plugin:1.0.0", "invoke_tool", argsJSON)
	if err != nil {
		t.Fatalf("NewInvokeContext() error: %v", err)
	}
	if ctx.PluginID != "author/plugin:1.0.0" {
		t.Errorf("PluginID = %q; want %q", ctx.PluginID, "author/plugin:1.0.0")
	}
	if ctx.Action != "invoke_tool" {
		t.Errorf("Action = %q; want %q", ctx.Action, "invoke_tool")
	}
	if ctx.Request.TenantID != "t1" {
		t.Errorf("TenantID = %q; want %q", ctx.Request.TenantID, "t1")
	}
	if ctx.Request.UserID != "u1" {
		t.Errorf("UserID = %q; want %q", ctx.Request.UserID, "u1")
	}
}

func TestNewInvokeContext_DefaultTenantID(t *testing.T) {
	argsJSON := `{"data":{"key":"val"}}`
	ctx, err := NewInvokeContext("plugin", "invoke_tool", argsJSON)
	if err != nil {
		t.Fatalf("NewInvokeContext() error: %v", err)
	}
	if ctx.Request.TenantID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("TenantID = %q; want nil UUID", ctx.Request.TenantID)
	}
}

func TestNewInvokeContext_InvalidJSON(t *testing.T) {
	_, err := NewInvokeContext("plugin", "invoke_tool", "not-json")
	if err == nil {
		t.Fatal("NewInvokeContext() should fail on invalid JSON")
	}
	se, ok := err.(*SlimError)
	if !ok {
		t.Fatalf("expected *SlimError, got %T", err)
	}
	if se.Code != ErrInvalidArgsJSON {
		t.Errorf("Code = %q; want %q", se.Code, ErrInvalidArgsJSON)
	}
}

func TestLoadConfigFromFile_Valid(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	content := `{
		"mode": "remote",
		"remote": {
			"daemon_addr": "http://localhost:5003",
			"daemon_key": "testkey"
		}
	}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	cfg, err := LoadConfigFromFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfigFromFile() error: %v", err)
	}
	if cfg.Mode != ModeRemote {
		t.Errorf("Mode = %q; want %q", cfg.Mode, ModeRemote)
	}
	if cfg.Remote.DaemonAddr != "http://localhost:5003" {
		t.Errorf("DaemonAddr = %q; want %q", cfg.Remote.DaemonAddr, "http://localhost:5003")
	}
}

func TestLoadConfigFromFile_NotFound(t *testing.T) {
	_, err := LoadConfigFromFile("/nonexistent/config.json")
	if err == nil {
		t.Fatal("LoadConfigFromFile() should fail on missing file")
	}
}

func TestLoadConfigFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(cfgPath, []byte("{invalid"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	_, err := LoadConfigFromFile(cfgPath)
	if err == nil {
		t.Fatal("LoadConfigFromFile() should fail on invalid JSON")
	}
}

func TestLoadConfigFromFile_ValidationFails(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	content := `{"mode": "remote", "remote": {}}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	_, err := LoadConfigFromFile(cfgPath)
	if err == nil {
		t.Fatal("LoadConfigFromFile() should fail when remote config is incomplete")
	}
}

func TestLoadExtractConfig_LocalPathDoesNotRequireFolder(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	content := `{"mode": "local"}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	cfg, err := LoadExtractConfig(cfgPath, true)
	if err != nil {
		t.Fatalf("LoadExtractConfig() error: %v", err)
	}
	if cfg.Mode != ModeLocal {
		t.Fatalf("Mode = %q; want %q", cfg.Mode, ModeLocal)
	}
}

func TestLoadExtractConfig_LocalIDRequiresFolder(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	content := `{"mode": "local"}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := LoadExtractConfig(cfgPath, false)
	if err == nil {
		t.Fatal("LoadExtractConfig() should fail when local extract uses -id without folder")
	}
	se, ok := err.(*SlimError)
	if !ok {
		t.Fatalf("expected *SlimError, got %T", err)
	}
	if se.Code != ErrConfigInvalid {
		t.Fatalf("Code = %q; want %q", se.Code, ErrConfigInvalid)
	}
}
