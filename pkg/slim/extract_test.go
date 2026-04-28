package slim

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestExtractLocal_FromDirectory(t *testing.T) {
	fixture := filepath.Join("..", "..", "internal", "core", "local_runtime", "testdata", "plugin-without-dependencies")

	result, err := ExtractLocal(ExtractOptions{Path: fixture}, &LocalConfig{})
	if err != nil {
		t.Fatalf("ExtractLocal() error: %v", err)
	}

	if result.Manifest.Name != "no-deps-test-plugin" {
		t.Fatalf("manifest name = %q; want %q", result.Manifest.Name, "no-deps-test-plugin")
	}
	if result.Manifest.Tool == nil {
		t.Fatal("manifest tool declaration is nil")
	}
	if len(result.Manifest.Tool.Tools) != 1 {
		t.Fatalf("tool count = %d; want 1", len(result.Manifest.Tool.Tools))
	}
}

func TestExtractLocal_FromPackage(t *testing.T) {
	fixture := filepath.Join("..", "..", "internal", "core", "local_runtime", "testdata", "plugin-without-dependencies")
	pkgPath := zipFixture(t, fixture)

	result, err := ExtractLocal(ExtractOptions{Path: pkgPath}, &LocalConfig{})
	if err != nil {
		t.Fatalf("ExtractLocal() error: %v", err)
	}

	if result.Manifest.Name != "no-deps-test-plugin" {
		t.Fatalf("manifest name = %q; want %q", result.Manifest.Name, "no-deps-test-plugin")
	}
	if result.UniqueIdentifier == "" {
		t.Fatal("unique identifier is empty")
	}
}

func TestExtractLocal_Validation(t *testing.T) {
	_, err := ExtractLocal(ExtractOptions{}, &LocalConfig{})
	assertSlimErrorCode(t, err, ErrInvalidInput)

	_, err = ExtractLocal(ExtractOptions{PluginID: "plugin", Path: "plugin.difypkg"}, &LocalConfig{})
	assertSlimErrorCode(t, err, ErrInvalidInput)

	_, err = ExtractLocal(ExtractOptions{PluginID: "plugin"}, &LocalConfig{})
	assertSlimErrorCode(t, err, ErrConfigInvalid)
}

func TestRunExtract_UnsupportedOutput(t *testing.T) {
	err := RunExtract(&SlimConfig{Mode: ModeLocal}, ExtractOptions{Output: "yaml"}, &bytes.Buffer{})
	assertSlimErrorCode(t, err, ErrInvalidInput)
}

func TestRunExtract_OutputsSlimArgsEnvelope(t *testing.T) {
	fixture := filepath.Join("..", "..", "internal", "core", "local_runtime", "testdata", "plugin-without-dependencies")
	var buf bytes.Buffer

	err := RunExtract(
		&SlimConfig{Mode: ModeLocal},
		ExtractOptions{Path: fixture, Output: OutputJSON},
		&buf,
	)
	if err != nil {
		t.Fatalf("RunExtract() error: %v", err)
	}

	ctx, err := NewInvokeContext("test/no-deps-test-plugin:0.0.1@checksum", "invoke_tool", buf.String())
	if err != nil {
		t.Fatalf("NewInvokeContext() error: %v", err)
	}
	if len(ctx.Request.Data) == 0 {
		t.Fatal("request data is empty")
	}

	if _, _, err := TransformRequest(ctx); err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	var args struct {
		TenantID string         `json:"tenant_id"`
		Data     map[string]any `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &args); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if args.TenantID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("tenant_id = %q; want nil UUID", args.TenantID)
	}
	manifest, ok := args.Data["manifest"].(map[string]any)
	if !ok {
		t.Fatalf("data.manifest is %T; want object", args.Data["manifest"])
	}
	if manifest["name"] != "no-deps-test-plugin" {
		t.Fatalf("manifest name = %q; want no-deps-test-plugin", manifest["name"])
	}
	if _, ok := manifest["label"]; ok {
		t.Fatal("manifest label should be removed by extract cleanup")
	}
	if _, ok := manifest["description"]; ok {
		t.Fatal("manifest description should be removed by extract cleanup")
	}
}

func TestRunExtract_ActionArgsAreInvokableBySlim(t *testing.T) {
	fixture := filepath.Join("..", "..", "internal", "core", "local_runtime", "testdata", "plugin-without-dependencies")
	var buf bytes.Buffer

	err := RunExtract(
		&SlimConfig{Mode: ModeLocal},
		ExtractOptions{Path: fixture, Action: "invoke_tool", Output: OutputJSON},
		&buf,
	)
	if err != nil {
		t.Fatalf("RunExtract() error: %v", err)
	}

	ctx, err := NewInvokeContext("test/no-deps-test-plugin:0.0.1@checksum", "invoke_tool", buf.String())
	if err != nil {
		t.Fatalf("NewInvokeContext() error: %v", err)
	}
	transformed, _, err := TransformRequest(ctx)
	if err != nil {
		t.Fatalf("TransformRequest() error: %v", err)
	}

	var msg map[string]any
	if err := json.Unmarshal(transformed, &msg); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	data := msg["data"].(map[string]any)
	if data["provider"] != "test_provider" {
		t.Fatalf("provider = %q; want test_provider", data["provider"])
	}
	if data["tool"] != "test_tool" {
		t.Fatalf("tool = %q; want test_tool", data["tool"])
	}
}

func TestDaemonClient_Extract(t *testing.T) {
	var gotAPIKey string
	var gotPluginID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("X-Api-Key")
		gotPluginID = r.Header.Get("X-Plugin-Unique-Identifier")
		if r.URL.Path != extractPath {
			t.Fatalf("path = %q; want %q", r.URL.Path, extractPath)
		}

		manifest := plugin_entities.PluginDeclaration{}
		manifest.Name = "remote-plugin"
		json.NewEncoder(w).Encode(daemonExtractResponse{
			Code:    0,
			Message: "success",
			Data: ExtractResult{
				UniqueIdentifier: "author/remote-plugin:0.0.1@checksum",
				Manifest:         manifest,
			},
		})
	}))
	defer srv.Close()

	client := NewDaemonClient(srv.URL, "secret")
	result, err := client.Extract("author/remote-plugin:0.0.1@checksum")
	if err != nil {
		t.Fatalf("Extract() error: %v", err)
	}

	if gotAPIKey != "secret" {
		t.Fatalf("X-Api-Key = %q; want secret", gotAPIKey)
	}
	if gotPluginID != "author/remote-plugin:0.0.1@checksum" {
		t.Fatalf("X-Plugin-Unique-Identifier = %q; want plugin id", gotPluginID)
	}
	if result.Manifest.Name != "remote-plugin" {
		t.Fatalf("manifest name = %q; want remote-plugin", result.Manifest.Name)
	}
}

func TestExtractRemote_Validation(t *testing.T) {
	_, err := ExtractRemote(ExtractOptions{Path: "./plugin"}, &RemoteConfig{})
	assertSlimErrorCode(t, err, ErrInvalidInput)

	_, err = ExtractRemote(ExtractOptions{}, &RemoteConfig{})
	assertSlimErrorCode(t, err, ErrInvalidInput)
}

func zipFixture(t *testing.T, src string) string {
	t.Helper()

	dst := filepath.Join(t.TempDir(), "plugin.difypkg")
	file, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})
	if err != nil {
		t.Fatalf("write zip: %v", err)
	}

	return dst
}

func assertSlimErrorCode(t *testing.T, err error, code ErrorCode) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	se, ok := err.(*SlimError)
	if !ok {
		t.Fatalf("expected *SlimError, got %T: %v", err, err)
	}
	if se.Code != code {
		t.Fatalf("error code = %s; want %s; message=%s", se.Code, code, se.Message)
	}
	if strings.TrimSpace(se.Message) == "" {
		t.Fatal("error message is empty")
	}
}
