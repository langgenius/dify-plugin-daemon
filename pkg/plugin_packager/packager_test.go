package plugin_packager

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/signer"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/signer/withkey"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
)

//go:embed testdata/manifest.yaml
var manifest []byte

//go:embed testdata/neko.yaml
var neko []byte

//go:embed testdata/.difyignore
var dify_ignore []byte

//go:embed testdata/ignored
var ignored []byte

//go:embed testdata/_assets/test.svg
var test_svg []byte

//go:embed testdata/keys
var keys embed.FS

// createMinimalPlugin creates a minimal test plugin and returns the zip file
func createMinimalPlugin(t *testing.T) []byte {
	// create a temp directory
	tempDir := t.TempDir()

	// create basic files
	if err := os.WriteFile(filepath.Join(tempDir, "manifest.yaml"), manifest, 0644); err != nil {
		t.Errorf("failed to write manifest: %s", err.Error())
		return nil
	}

	if err := os.WriteFile(filepath.Join(tempDir, "neko.yaml"), neko, 0644); err != nil {
		t.Errorf("failed to write neko: %s", err.Error())
		return nil
	}

	// create _assets directory and files
	if err := os.MkdirAll(filepath.Join(tempDir, "_assets"), 0755); err != nil {
		t.Errorf("failed to create _assets directory: %s", err.Error())
		return nil
	}

	if err := os.WriteFile(filepath.Join(tempDir, "_assets/test.svg"), test_svg, 0644); err != nil {
		t.Errorf("failed to write test.svg: %s", err.Error())
		return nil
	}

	// create decoder
	originDecoder, err := decoder.NewFSPluginDecoder(tempDir)
	if err != nil {
		t.Errorf("failed to create decoder: %s", err.Error())
		return nil
	}

	// create packager
	packager := packager.NewPackager(originDecoder)

	// pack
	zip, err := packager.Pack(52428800)
	if err != nil {
		t.Errorf("failed to pack: %s", err.Error())
		return nil
	}

	return zip
}

func TestPackagerAndVerifier(t *testing.T) {
	// create a temp directory
	tempDir := t.TempDir()

	// create manifest
	if err := os.WriteFile(filepath.Join(tempDir, "manifest.yaml"), manifest, 0644); err != nil {
		t.Errorf("failed to write manifest: %s", err.Error())
		return
	}

	if err := os.WriteFile(filepath.Join(tempDir, "neko.yaml"), neko, 0644); err != nil {
		t.Errorf("failed to write neko: %s", err.Error())
		return
	}

	// create .difyignore
	if err := os.WriteFile(filepath.Join(tempDir, ".difyignore"), dify_ignore, 0644); err != nil {
		t.Errorf("failed to write .difyignore: %s", err.Error())
		return
	}

	// create ignored
	if err := os.WriteFile(filepath.Join(tempDir, "ignored"), ignored, 0644); err != nil {
		t.Errorf("failed to write ignored: %s", err.Error())
		return
	}

	// create ignored_paths
	if err := os.MkdirAll(filepath.Join(tempDir, "ignored_paths"), 0755); err != nil {
		t.Errorf("failed to create ignored_paths directory: %s", err.Error())
		return
	}

	// create ignored_paths/ignored
	if err := os.WriteFile(filepath.Join(tempDir, "ignored_paths/ignored"), ignored, 0644); err != nil {
		t.Errorf("failed to write ignored_paths/ignored: %s", err.Error())
		return
	}

	if err := os.MkdirAll(filepath.Join(tempDir, "_assets"), 0755); err != nil {
		t.Errorf("failed to create _assets directory: %s", err.Error())
		return
	}

	if err := os.WriteFile(filepath.Join(tempDir, "_assets/test.svg"), test_svg, 0644); err != nil {
		t.Errorf("failed to write test.svg: %s", err.Error())
		return
	}

	originDecoder, err := decoder.NewFSPluginDecoder(tempDir)
	if err != nil {
		t.Errorf("failed to create decoder: %s", err.Error())
		return
	}

	// walk
	err = originDecoder.Walk(func(filename string, dir string) error {
		if filename == "ignored" {
			return fmt.Errorf("should not walk into ignored")
		}
		if strings.HasPrefix(filename, "ignored_paths") {
			return fmt.Errorf("should not walk into ignored_paths")
		}
		return nil
	})
	if err != nil {
		t.Errorf("failed to walk: %s", err.Error())
		return
	}

	// check assets
	assets, err := originDecoder.Assets()
	if err != nil {
		t.Errorf("failed to get assets: %s", err.Error())
		return
	}

	if assets["test.svg"] == nil {
		t.Errorf("should have test.svg asset, got %v", assets)
		return
	}

	packager := packager.NewPackager(originDecoder)

	// pack
	zip, err := packager.Pack(52428800)
	if err != nil {
		t.Errorf("failed to pack: %s", err.Error())
		return
	}

	// sign
	signed, err := signer.SignPlugin(zip, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	})
	if err != nil {
		t.Errorf("failed to sign: %s", err.Error())
		return
	}

	signedDecoder, err := decoder.NewZipPluginDecoder(signed)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// check assets
	assets, err = signedDecoder.Assets()
	if err != nil {
		t.Errorf("failed to get assets: %s", err.Error())
		return
	}

	if assets["test.svg"] == nil {
		t.Errorf("should have test.svg asset, got %v", assets)
		return
	}

	// verify
	err = decoder.VerifyPlugin(signedDecoder)
	if err != nil {
		t.Errorf("failed to verify: %s", err.Error())
		return
	}
}

func TestWrongSign(t *testing.T) {
	// create a minimal test plugin
	zip := createMinimalPlugin(t)
	if zip == nil {
		return
	}

	// sign
	signed, err := signer.SignPlugin(zip, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	})
	if err != nil {
		t.Errorf("failed to sign: %s", err.Error())
		return
	}

	// modify the signed file, signature is at the end of the file
	signed[len(signed)-1] = 0
	signed[len(signed)-2] = 0

	// create a new decoder
	signedDecoder, err := decoder.NewZipPluginDecoder(signed)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// verify (expected to fail)
	err = decoder.VerifyPlugin(signedDecoder)
	if err == nil {
		t.Errorf("should fail to verify")
		return
	}
}

// loadPublicKeyFile loads a key file from the embed.FS and returns the public key
func loadPublicKeyFile(t *testing.T, keyFile string) *rsa.PublicKey {
	keyBytes, err := keys.ReadFile(filepath.Join("testdata/keys", keyFile))
	if err != nil {
		t.Fatalf("failed to read key file: %s", err.Error())
	}
	key, err := encryption.LoadPublicKey(keyBytes)
	if err != nil {
		t.Fatalf("failed to load public key: %s", err.Error())
	}
	return key
}

// loadPrivateKeyFile loads a key file from the embed.FS and returns the private key
func loadPrivateKeyFile(t *testing.T, keyFile string) *rsa.PrivateKey {
	keyBytes, err := keys.ReadFile(filepath.Join("testdata/keys", keyFile))
	if err != nil {
		t.Fatalf("failed to read key file: %s", err.Error())
	}
	key, err := encryption.LoadPrivateKey(keyBytes)
	if err != nil {
		t.Fatalf("failed to load private key: %s", err.Error())
	}
	return key
}

func TestPackager_ExportsRequirementsFromUvLock(t *testing.T) {
	tempDir := t.TempDir()
	writePythonPluginFiles(t, tempDir, false)
	writeFakeUV(t, tempDir, "dify-plugin==0.0.1\nrequests==2.32.3\n")

	originDecoder, err := decoder.NewFSPluginDecoder(tempDir)
	if err != nil {
		t.Fatalf("failed to create decoder: %v", err)
	}

	p := packager.NewPackager(originDecoder)
	zipBytes, err := p.Pack(52428800)
	if err != nil {
		t.Fatalf("failed to pack: %v", err)
	}

	files := unzipEntries(t, zipBytes)
	if _, ok := files["pyproject.toml"]; !ok {
		t.Fatal("expected pyproject.toml in package")
	}
	if _, ok := files["uv.lock"]; !ok {
		t.Fatal("expected uv.lock in package")
	}
	got, ok := files["requirements.txt"]
	if !ok {
		t.Fatal("expected generated requirements.txt in package")
	}
	if string(got) != "dify-plugin==0.0.1\nrequests==2.32.3\n" {
		t.Fatalf("unexpected generated requirements.txt: %q", string(got))
	}
	if _, err := os.Stat(filepath.Join(tempDir, "requirements.txt")); !os.IsNotExist(err) {
		t.Fatalf("requirements.txt should not be created in source dir, stat err=%v", err)
	}
}

func TestPackager_DoesNotRegenerateExistingRequirements(t *testing.T) {
	tempDir := t.TempDir()
	writePythonPluginFiles(t, tempDir, true)
	writeFakeUV(t, tempDir, "generated-should-not-be-used==1.0.0\n")

	originDecoder, err := decoder.NewFSPluginDecoder(tempDir)
	if err != nil {
		t.Fatalf("failed to create decoder: %v", err)
	}

	p := packager.NewPackager(originDecoder)
	zipBytes, err := p.Pack(52428800)
	if err != nil {
		t.Fatalf("failed to pack: %v", err)
	}

	files := unzipEntries(t, zipBytes)
	got, ok := files["requirements.txt"]
	if !ok {
		t.Fatal("expected existing requirements.txt in package")
	}
	if string(got) != "existing-package==1.0.0\n" {
		t.Fatalf("existing requirements.txt should be preserved, got %q", string(got))
	}
}

func TestPackager_PythonTemplateDifyignoreKeepsUvLock(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "cmd", "commandline", "plugin", "templates", "python", ".difyignore"))
	if err != nil {
		t.Fatalf("failed to read template .difyignore: %v", err)
	}
	if strings.Contains(string(content), "\nuv.lock\n") {
		t.Fatal("template .difyignore should not exclude uv.lock")
	}
}

func TestPackager_PythonTemplateGitignoreKeepsUvLock(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "cmd", "commandline", "plugin", "templates", "python", ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read template .gitignore: %v", err)
	}
	if strings.Contains(string(content), "\nuv.lock\n") {
		t.Fatal("template .gitignore should not exclude uv.lock")
	}
}

func writePythonPluginFiles(t *testing.T, root string, withRequirements bool) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, "manifest.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "neko.yaml"), neko, 0o644); err != nil {
		t.Fatalf("failed to write neko.yaml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "_assets"), 0o755); err != nil {
		t.Fatalf("failed to create _assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "_assets", "test.svg"), test_svg, 0o644); err != nil {
		t.Fatalf("failed to write test.svg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\nname = \"python-packager-test\"\nversion = \"1.0.0\"\ndependencies = [\"requests==2.32.3\"]\n"), 0o644); err != nil {
		t.Fatalf("failed to write pyproject.toml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "uv.lock"), []byte("version = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write uv.lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte("print('hello')\n"), 0o644); err != nil {
		t.Fatalf("failed to write main.py: %v", err)
	}
	if withRequirements {
		if err := os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("existing-package==1.0.0\n"), 0o644); err != nil {
			t.Fatalf("failed to write requirements.txt: %v", err)
		}
	}
}

func writeFakeUV(t *testing.T, root string, output string) {
	t.Helper()

	scriptPath := filepath.Join(root, "fake-uv")
	script := "#!/bin/sh\ncat <<'EOF'\n" + output + "EOF\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake uv: %v", err)
	}
	t.Setenv("UV_PATH", scriptPath)
}

func unzipEntries(t *testing.T, data []byte) map[string][]byte {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}

	files := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("failed to open zip entry %s: %v", file.Name, err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("failed to read zip entry %s: %v", file.Name, err)
		}
		files[file.Name] = content
	}
	return files
}

// extractPublicKey extracts the key file from the embed.FS and returns the file path
func extractKeyFile(t *testing.T, keyFile string, tmpDir string) string {
	keyBytes, err := keys.ReadFile(filepath.Join("testdata/keys", keyFile))
	if err != nil {
		t.Fatalf("failed to read key file: %s", err.Error())
	}
	keyPath := filepath.Join(tmpDir, keyFile)
	if err := os.WriteFile(keyPath, keyBytes, 0644); err != nil {
		t.Fatalf("failed to write key file: %s", err.Error())
	}
	return keyPath
}

func TestSignPluginWithPrivateKey(t *testing.T) {
	// load public keys from embed.FS
	publicKey1 := loadPublicKeyFile(t, "test_key_pair_1.public.pem")
	publicKey2 := loadPublicKeyFile(t, "test_key_pair_2.public.pem")

	// load private keys from embed.FS
	privateKey1 := loadPrivateKeyFile(t, "test_key_pair_1.private.pem")
	privateKey2 := loadPrivateKeyFile(t, "test_key_pair_2.private.pem")

	// create a minimal test plugin
	zip := createMinimalPlugin(t)
	if zip == nil {
		return
	}

	// sign with private key 1 and create decoder
	signed1, err := withkey.SignPluginWithPrivateKey(zip, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	}, privateKey1)
	if err != nil {
		t.Errorf("failed to sign with private key 1: %s", err.Error())
		return
	}
	signedDecoder1, err := decoder.NewZipPluginDecoder(signed1)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// sign with private key 2 and create decoder
	signed2, err := withkey.SignPluginWithPrivateKey(zip, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	}, privateKey2)
	if err != nil {
		t.Errorf("failed to sign with private key 2: %s", err.Error())
		return
	}
	signedDecoder2, err := decoder.NewZipPluginDecoder(signed2)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// tamper the signed1 file and create decoder
	modifiedSigned1 := make([]byte, len(signed1))
	copy(modifiedSigned1, signed1)
	modifiedSigned1[len(modifiedSigned1)-10] = 0
	modifiedDecoder1, err := decoder.NewZipPluginDecoder(modifiedSigned1)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// define test cases
	tests := []struct {
		name          string
		signedDecoder decoder.PluginDecoder
		publicKeys    []*rsa.PublicKey
		expectSuccess bool
	}{
		{
			name:          "verify plugin signed with private key 1 using embedded public key (should fail)",
			signedDecoder: signedDecoder1,
			publicKeys:    nil, // use embedded public key
			expectSuccess: false,
		},
		{
			name:          "verify plugin signed with private key 1 using public key 1 (should succeed)",
			signedDecoder: signedDecoder1,
			publicKeys:    []*rsa.PublicKey{publicKey1},
			expectSuccess: true,
		},
		{
			name:          "verify plugin signed with private key 1 using public key 2 (should fail)",
			signedDecoder: signedDecoder1,
			publicKeys:    []*rsa.PublicKey{publicKey2},
			expectSuccess: false,
		},
		{
			name:          "verify plugin signed with private key 2 using public key 1 (should fail)",
			signedDecoder: signedDecoder2,
			publicKeys:    []*rsa.PublicKey{publicKey1},
			expectSuccess: false,
		},
		{
			name:          "verify plugin signed with private key 2 using public key 2 (should succeed)",
			signedDecoder: signedDecoder2,
			publicKeys:    []*rsa.PublicKey{publicKey2},
			expectSuccess: true,
		},
		{
			name:          "verify modified plugin signed with private key 1 using public key 1 (should fail)",
			signedDecoder: modifiedDecoder1,
			publicKeys:    []*rsa.PublicKey{publicKey1},
			expectSuccess: false,
		},
		{
			name:          "verify modified plugin signed with private key 1 using public key 2 (should fail)",
			signedDecoder: modifiedDecoder1,
			publicKeys:    []*rsa.PublicKey{publicKey2},
			expectSuccess: false,
		},
	}

	// run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.publicKeys == nil {
				err = decoder.VerifyPlugin(tt.signedDecoder)
			} else {
				err = decoder.VerifyPluginWithPublicKeys(tt.signedDecoder, tt.publicKeys)
			}

			if tt.expectSuccess && err != nil {
				t.Errorf("expected success but got error: %s", err.Error())
			}
			if !tt.expectSuccess && err == nil {
				t.Errorf("expected failure but got success")
			}
		})
	}
}

func TestVerifyPluginWithThirdPartyKeys(t *testing.T) {
	// create a temporary directory for the public key files (needed for storing the paths in environment variable)
	tempDir := t.TempDir()

	// extract public keys to files from embed.FS (needed for storing the paths in environment variable)
	publicKey1Path := extractKeyFile(t, "test_key_pair_1.public.pem", tempDir)
	publicKey2Path := extractKeyFile(t, "test_key_pair_2.public.pem", tempDir)

	// load private keys from embed.FS
	privateKey1 := loadPrivateKeyFile(t, "test_key_pair_1.private.pem")
	privateKey2 := loadPrivateKeyFile(t, "test_key_pair_2.private.pem")

	// create a minimal test plugin
	zip := createMinimalPlugin(t)
	if zip == nil {
		return
	}

	// sign with private key 1 and create decoder
	signed1, err := withkey.SignPluginWithPrivateKey(zip, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	}, privateKey1)
	if err != nil {
		t.Errorf("failed to sign with private key 1: %s", err.Error())
		return
	}
	signedDecoder1, err := decoder.NewZipPluginDecoder(signed1)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// sign with private key 2 and create decoder
	signed2, err := withkey.SignPluginWithPrivateKey(zip, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	}, privateKey2)
	if err != nil {
		t.Errorf("failed to sign with private key 2: %s", err.Error())
		return
	}
	signedDecoder2, err := decoder.NewZipPluginDecoder(signed2)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// tamper the signed1 file and create decoder
	modifiedSigned1 := make([]byte, len(signed1))
	copy(modifiedSigned1, signed1)
	modifiedSigned1[len(modifiedSigned1)-10] = 0
	modifiedDecoder1, err := decoder.NewZipPluginDecoder(modifiedSigned1)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// define test cases
	tests := []struct {
		name          string
		keyPaths      string
		signedDecoder decoder.PluginDecoder
		expectSuccess bool
	}{
		{
			name:          "third-party verification with public key 1 (should succeed)",
			keyPaths:      publicKey1Path,
			signedDecoder: signedDecoder1,
			expectSuccess: true,
		},
		{
			name:          "third-party verification with public key 2 (should fail)",
			keyPaths:      publicKey2Path,
			signedDecoder: signedDecoder1,
			expectSuccess: false,
		},
		{
			name:          "third-party verification with both keys (should succeed)",
			keyPaths:      fmt.Sprintf("%s,%s", publicKey1Path, publicKey2Path),
			signedDecoder: signedDecoder1,
			expectSuccess: true,
		},
		{
			name:          "third-party verification with empty key path (should fail)",
			keyPaths:      "",
			signedDecoder: signedDecoder1,
			expectSuccess: false,
		},
		{
			name:          "third-party verification with non-existent key path (should fail)",
			keyPaths:      "/non/existent/path.pem",
			signedDecoder: signedDecoder1,
			expectSuccess: false,
		},
		{
			name:          "third-party verification with multiple keys including non-existent path (should fail)",
			keyPaths:      fmt.Sprintf("%s,%s,/non/existent/path.pem", publicKey1Path, publicKey2Path),
			signedDecoder: signedDecoder1,
			expectSuccess: false,
		},
		{
			name:          "third-party verification with multiple keys including extra spaces (should succeed)",
			keyPaths:      fmt.Sprintf(" %s , %s ", publicKey1Path, publicKey2Path),
			signedDecoder: signedDecoder1,
			expectSuccess: true,
		},
		{
			name:          "third-party verification with both keys, for file signed with key 2 (should succeed)",
			keyPaths:      fmt.Sprintf("%s,%s", publicKey1Path, publicKey2Path),
			signedDecoder: signedDecoder2,
			expectSuccess: true,
		},
		{
			name:          "third-party verification with both keys, for modified file (should fail)",
			keyPaths:      fmt.Sprintf("%s,%s", publicKey1Path, publicKey2Path),
			signedDecoder: modifiedDecoder1,
			expectSuccess: false,
		},
	}

	// run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decoder.VerifyPluginWithPublicKeyPaths(tt.signedDecoder, strings.Split(tt.keyPaths, ","))
			if tt.expectSuccess && err != nil {
				t.Errorf("expected success but got error: %s", err.Error())
			}
			if !tt.expectSuccess && err == nil {
				t.Errorf("expected failure but got success")
			}
		})
	}
}
