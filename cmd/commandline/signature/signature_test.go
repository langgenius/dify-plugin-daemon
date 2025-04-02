package signature

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/encryption"
)

//go:embed testdata/dummy_plugin.difypkg
var dummyPlugin []byte

func TestGenerateKeyPair(t *testing.T) {
	// create a temporary directory for testing
	tempDir := t.TempDir()

	// create a key pair
	keyPairName := filepath.Join(tempDir, "test_key_pair")
	GenerateKeyPair(keyPairName)
	privateKeyPath := keyPairName + ".private.pem"
	publicKeyPath := keyPairName + ".public.pem"

	// check if the key files are created
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Errorf("Private key file was not created: %s", privateKeyPath)
	}
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Errorf("Public key file was not created: %s", publicKeyPath)
	}

	// check if the key files can be loaded
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to read private key file: %v", err)
	}
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to read public key file: %v", err)
	}

	// check if the keys can be loaded
	_, err = encryption.LoadPrivateKey(privateKeyBytes)
	if err != nil {
		t.Errorf("Failed to load private key: %v", err)
	}
	_, err = encryption.LoadPublicKey(publicKeyBytes)
	if err != nil {
		t.Errorf("Failed to load public key: %v", err)
	}
}

func TestSignAndVerify(t *testing.T) {
	// create a temporary directory for testing
	tempDir := t.TempDir()

	// extract the minimal plugin content from the embedded data to a file
	dummyPluginPath := filepath.Join(tempDir, "dummy_plugin.difypkg")
	if err := os.WriteFile(dummyPluginPath, dummyPlugin, 0644); err != nil {
		t.Fatalf("Failed to create dummy plugin file: %v", err)
	}

	// create two key pairs
	keyPair1Name := filepath.Join(tempDir, "test_key_pair_1")
	keyPair2Name := filepath.Join(tempDir, "test_key_pair_2")
	GenerateKeyPair(keyPair1Name)
	GenerateKeyPair(keyPair2Name)
	privateKey1Path := keyPair1Name + ".private.pem"
	publicKey1Path := keyPair1Name + ".public.pem"
	privateKey2Path := keyPair2Name + ".private.pem"
	publicKey2Path := keyPair2Name + ".public.pem"

	// test case definition for table-driven tests
	type testCase struct {
		name          string
		signKeyPath   string
		verifyKeyPath string
		expectSuccess bool
	}

	// test cases
	tests := []testCase{
		{
			name:          "sign with keypair1, verify with keypair1",
			signKeyPath:   privateKey1Path,
			verifyKeyPath: publicKey1Path,
			expectSuccess: true,
		},
		{
			name:          "sign with keypair1, verify with keypair2",
			signKeyPath:   privateKey1Path,
			verifyKeyPath: publicKey2Path,
			expectSuccess: false,
		},
		{
			name:          "sign with keypair2, verify with keypair2",
			signKeyPath:   privateKey2Path,
			verifyKeyPath: publicKey2Path,
			expectSuccess: true,
		},
		{
			name:          "sign with keypair2, verify with keypair1",
			signKeyPath:   privateKey2Path,
			verifyKeyPath: publicKey1Path,
			expectSuccess: false,
		},
		{
			name:          "sign with keypair1, verify without key",
			signKeyPath:   privateKey1Path,
			verifyKeyPath: "",
			expectSuccess: false,
		},
		{
			name:          "sign with keypair2, verify without key",
			signKeyPath:   privateKey2Path,
			verifyKeyPath: "",
			expectSuccess: false,
		},
	}

	// execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a temporary file for each test case
			testPluginPath := filepath.Join(tempDir, "test_plugin_"+tt.name+".difypkg")
			if err := os.WriteFile(testPluginPath, dummyPlugin, 0644); err != nil {
				t.Fatalf("Failed to create test plugin file: %v", err)
			}

			// sign the plugin
			Sign(testPluginPath, tt.signKeyPath)

			// get the path of the signed plugin
			dir := filepath.Dir(testPluginPath)
			base := filepath.Base(testPluginPath)
			ext := filepath.Ext(base)
			name := base[:len(base)-len(ext)]
			dummyPluginPath := filepath.Join(dir, name+".signed"+ext)

			// check if the signed plugin file was created
			if _, err := os.Stat(dummyPluginPath); os.IsNotExist(err) {
				t.Fatalf("Signed plugin file was not created: %s", dummyPluginPath)
			}

			// verify the signed plugin and check the result
			err := Verify(dummyPluginPath, tt.verifyKeyPath)
			if tt.expectSuccess && err != nil {
				t.Errorf("Expected verification to succeed, but got error: %v", err)
			} else if !tt.expectSuccess && err == nil {
				t.Errorf("Expected verification to fail, but it succeeded")
			}
		})
	}
}

// TestVerifyUnsigned tests verification of an unsigned difypkg file
func TestVerifyUnsigned(t *testing.T) {
	// create a temporary directory for testing
	tempDir := t.TempDir()

	// extract the minimal plugin content from the embedded data to a file
	dummyPluginPath := filepath.Join(tempDir, "dummy_plugin.difypkg")
	if err := os.WriteFile(dummyPluginPath, dummyPlugin, 0644); err != nil {
		t.Fatalf("Failed to create dummy plugin file: %v", err)
	}

	// create a key pair
	keyPairName := filepath.Join(tempDir, "test_key_pair")
	GenerateKeyPair(keyPairName)
	publicKeyPath := keyPairName + ".public.pem"

	// Try to verify the unsigned plugin file
	err := Verify(dummyPluginPath, publicKeyPath)
	if err == nil {
		t.Errorf("Expected verification of unsigned file to fail, but it succeeded")
	}
}

// TestVerifyTampered tests verification of a tampered signed difypkg file
func TestVerifyTampered(t *testing.T) {
	// create a temporary directory for testing
	tempDir := t.TempDir()

	// extract the minimal plugin content from the embedded data to a file
	dummyPluginPath := filepath.Join(tempDir, "dummy_plugin.difypkg")
	if err := os.WriteFile(dummyPluginPath, dummyPlugin, 0644); err != nil {
		t.Fatalf("Failed to create dummy plugin file: %v", err)
	}

	// create a key pair
	keyPairName := filepath.Join(tempDir, "test_key_pair")
	GenerateKeyPair(keyPairName)
	privateKeyPath := keyPairName + ".private.pem"
	publicKeyPath := keyPairName + ".public.pem"

	// Sign the plugin
	Sign(dummyPluginPath, privateKeyPath)

	// Get the path of the signed plugin
	dir := filepath.Dir(dummyPluginPath)
	base := filepath.Base(dummyPluginPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	signedDummyPluginPath := filepath.Join(dir, name+".signed"+ext)

	// Read the signed plugin
	signedPluginData, err := os.ReadFile(signedDummyPluginPath)
	if err != nil {
		t.Fatalf("Failed to read signed plugin file: %v", err)
	}

	// tamper the signed plugin data
	signedPluginData[len(signedPluginData)-10] = 0

	// write the tampered data back to the file
	tamperedPluginPath := filepath.Join(tempDir, "tampered_plugin.signed.difypkg")
	if err := os.WriteFile(tamperedPluginPath, signedPluginData, 0644); err != nil {
		t.Fatalf("Failed to write tampered plugin file: %v", err)
	}

	// try to verify the tampered plugin file
	err = Verify(tamperedPluginPath, publicKeyPath)
	if err == nil {
		t.Errorf("Expected verification of tampered file to fail, but it succeeded")
	}
}
