package plugin_packager

import (
	_ "embed"
	"os"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/signer"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/verifier"
)

//go:embed manifest.yaml
var manifest []byte

func TestPackagerAndVerifier(t *testing.T) {
	// create a temp directory
	if err := os.Mkdir("temp", 0755); err != nil {
		t.Errorf("failed to create temp directory: %s", err.Error())
		return
	}
	defer func() {
		os.RemoveAll("temp")
		os.Remove("temp")
	}()

	// create manifest
	if err := os.WriteFile("temp/manifest.yaml", manifest, 0644); err != nil {
		t.Errorf("failed to write manifest: %s", err.Error())
		return
	}

	origin_decoder, err := decoder.NewFSPluginDecoder("temp")
	if err != nil {
		t.Errorf("failed to create decoder: %s", err.Error())
		return
	}

	packager := packager.NewPackager(origin_decoder)

	// pack
	zip, err := packager.Pack()
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

	signed_decoder, err := decoder.NewZipPluginDecoder(signed)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// verify
	err = verifier.VerifyPlugin(signed_decoder)
	if err != nil {
		t.Errorf("failed to verify: %s", err.Error())
		return
	}
}

func TestWrongSign(t *testing.T) {
	// create a temp directory
	if err := os.Mkdir("temp", 0755); err != nil {
		t.Errorf("failed to create temp directory: %s", err.Error())
		return
	}
	defer func() {
		os.RemoveAll("temp")
		os.Remove("temp")
	}()

	// create manifest
	if err := os.WriteFile("temp/manifest.yaml", manifest, 0644); err != nil {
		t.Errorf("failed to write manifest: %s", err.Error())
		return
	}

	origin_decoder, err := decoder.NewFSPluginDecoder("temp")
	if err != nil {
		t.Errorf("failed to create decoder: %s", err.Error())
		return
	}

	packager := packager.NewPackager(origin_decoder)

	// pack
	zip, err := packager.Pack()
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

	// modify the signed file, signature is at the end of the file
	signed[len(signed)-1] = 0
	signed[len(signed)-2] = 0

	// create a new decoder
	signed_decoder, err := decoder.NewZipPluginDecoder(signed)
	if err != nil {
		t.Errorf("failed to create zip decoder: %s", err.Error())
		return
	}

	// verify
	err = verifier.VerifyPlugin(signed_decoder)
	if err == nil {
		t.Errorf("should fail to verify")
		return
	}
}
