package bundle_packager

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryZipBundlePackager_AssetsExtraction(t *testing.T) {
	// Create a test ZIP file with assets
	zipBuffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(zipBuffer)

	// Add manifest.yaml
	manifest, err := zipWriter.Create("manifest.yaml")
	assert.NoError(t, err)
	_, err = manifest.Write([]byte("version: 0.0.1\nname: test-bundle\n"))
	assert.NoError(t, err)

	// Add assets in _assets directory using forward slashes (ZIP standard)
	asset1, err := zipWriter.Create("_assets/icon.png")
	assert.NoError(t, err)
	_, err = asset1.Write([]byte("fake-png-data"))
	assert.NoError(t, err)

	asset2, err := zipWriter.Create("_assets/logo.svg")
	assert.NoError(t, err)
	_, err = asset2.Write([]byte("fake-svg-data"))
	assert.NoError(t, err)

	// Add nested asset
	asset3, err := zipWriter.Create("_assets/subdir/image.jpg")
	assert.NoError(t, err)
	_, err = asset3.Write([]byte("fake-jpg-data"))
	assert.NoError(t, err)

	err = zipWriter.Close()
	assert.NoError(t, err)

	// Test MemoryZipBundlePackager
	packager, err := NewMemoryZipBundlePackager(zipBuffer.Bytes())
	assert.NoError(t, err)
	assert.NotNil(t, packager)

	// Verify assets are correctly extracted
	assets, err := packager.Assets()
	assert.NoError(t, err)
	assert.Len(t, assets, 3)

	// Check each asset exists with correct content (without _assets prefix)
	assert.Contains(t, assets, "icon.png")
	assert.Equal(t, []byte("fake-png-data"), assets["icon.png"])

	assert.Contains(t, assets, "logo.svg")
	assert.Equal(t, []byte("fake-svg-data"), assets["logo.svg"])

	assert.Contains(t, assets, "subdir/image.jpg")
	assert.Equal(t, []byte("fake-jpg-data"), assets["subdir/image.jpg"])
}

func TestNewMemoryZipBundlePackager_NoAssets(t *testing.T) {
	// Create a test ZIP file without assets
	zipBuffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(zipBuffer)

	// Add manifest.yaml only
	manifest, err := zipWriter.Create("manifest.yaml")
	assert.NoError(t, err)
	_, err = manifest.Write([]byte("version: 0.0.1\nname: test-bundle\n"))
	assert.NoError(t, err)

	err = zipWriter.Close()
	assert.NoError(t, err)

	// Test MemoryZipBundlePackager
	packager, err := NewMemoryZipBundlePackager(zipBuffer.Bytes())
	assert.NoError(t, err)
	assert.NotNil(t, packager)

	// Verify no assets are extracted
	assets, err := packager.Assets()
	assert.NoError(t, err)
	assert.Len(t, assets, 0)
}

func TestNewMemoryZipBundlePackager_ReadFile(t *testing.T) {
	// Create a test ZIP file
	zipBuffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(zipBuffer)

	// Add manifest.yaml
	manifest, err := zipWriter.Create("manifest.yaml")
	assert.NoError(t, err)
	_, err = manifest.Write([]byte("version: 0.0.1\nname: test-bundle\n"))
	assert.NoError(t, err)

	// Add a non-asset file
	readme, err := zipWriter.Create("README.md")
	assert.NoError(t, err)
	_, err = readme.Write([]byte("# Test Bundle"))
	assert.NoError(t, err)

	err = zipWriter.Close()
	assert.NoError(t, err)

	// Test MemoryZipBundlePackager
	packager, err := NewMemoryZipBundlePackager(zipBuffer.Bytes())
	assert.NoError(t, err)
	assert.NotNil(t, packager)

	// Test reading file
	content, err := packager.ReadFile("README.md")
	assert.NoError(t, err)
	assert.Equal(t, []byte("# Test Bundle"), content)

	// Test reading non-existent file
	_, err = packager.ReadFile("nonexistent.md")
	assert.Error(t, err)
}
