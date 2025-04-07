package plugin_packager

import (
	"crypto/rsa"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/signer"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/signer/withkey"
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
	signed, err := signer.SignPlugin(zip)
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
	signed, err := signer.SignPlugin(zip)
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
	signed1, err := withkey.SignPluginWithPrivateKey(zip, privateKey1)
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
	signed2, err := withkey.SignPluginWithPrivateKey(zip, privateKey2)
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
	signed1, err := withkey.SignPluginWithPrivateKey(zip, privateKey1)
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
	signed2, err := withkey.SignPluginWithPrivateKey(zip, privateKey2)
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
