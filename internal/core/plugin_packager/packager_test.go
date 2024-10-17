package plugin_packager

import (
	_ "embed"
	"fmt"
	"os"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/packager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/signer"
)

//go:embed manifest.yaml
var manifest []byte

//go:embed neko.yaml
var neko []byte

//go:embed .difyignore
var dify_ignore []byte

//go:embed ignored
var ignored []byte

//go:embed _assets/test.svg
var test_svg []byte

func TestPackagerAndVerifier(t *testing.T) {
	// create a temp directory
	os.RemoveAll("temp")
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

	if err := os.WriteFile("temp/neko.yaml", neko, 0644); err != nil {
		t.Errorf("failed to write neko: %s", err.Error())
		return
	}

	// create .difyignore
	if err := os.WriteFile("temp/.difyignore", dify_ignore, 0644); err != nil {
		t.Errorf("failed to write .difyignore: %s", err.Error())
		return
	}

	// create ignored
	if err := os.WriteFile("temp/ignored", ignored, 0644); err != nil {
		t.Errorf("failed to write ignored: %s", err.Error())
		return
	}

	if err := os.MkdirAll("temp/_assets", 0755); err != nil {
		t.Errorf("failed to create _assets directory: %s", err.Error())
		return
	}

	if err := os.WriteFile("temp/_assets/test.svg", test_svg, 0644); err != nil {
		t.Errorf("failed to write test.svg: %s", err.Error())
		return
	}

	origin_decoder, err := decoder.NewFSPluginDecoder("temp")
	if err != nil {
		t.Errorf("failed to create decoder: %s", err.Error())
		return
	}

	// walk
	err = origin_decoder.Walk(func(filename string, dir string) error {
		if filename == "ignored" {
			return fmt.Errorf("should not walk into ignored")
		}
		return nil
	})
	if err != nil {
		t.Errorf("failed to walk: %s", err.Error())
		return
	}

	// check assets
	assets, err := origin_decoder.Assets()
	if err != nil {
		t.Errorf("failed to get assets: %s", err.Error())
		return
	}

	if assets["test.svg"] == nil {
		t.Errorf("should have test.svg asset, got %v", assets)
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

	// check assets
	assets, err = signed_decoder.Assets()
	if err != nil {
		t.Errorf("failed to get assets: %s", err.Error())
		return
	}

	if assets["test.svg"] == nil {
		t.Errorf("should have test.svg asset, got %v", assets)
		return
	}

	// verify
	err = decoder.VerifyPlugin(signed_decoder)
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

	if err := os.WriteFile("temp/neko.yaml", neko, 0644); err != nil {
		t.Errorf("failed to write neko: %s", err.Error())
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
	err = decoder.VerifyPlugin(signed_decoder)
	if err == nil {
		t.Errorf("should fail to verify")
		return
	}
}
