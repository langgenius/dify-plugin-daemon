package decoder

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/assert"
)

type UnixPluginDecoder struct {
	PluginDecoder
	PluginDecoderHelper
}

func (d *UnixPluginDecoder) ReadFile(filename string) ([]byte, error) {
	return []byte("test"), nil
}

func (d *UnixPluginDecoder) ReadDir(dirname string) ([]string, error) {
	if dirname == "_assets" {
		return []string{
			"_assets/test.txt",
			"_assets/test2.txt",
		}, nil
	} else if dirname == "readme" {
		return []string{
			"readme/README_zh_Hans.md",
		}, nil
	}
	return nil, nil
}

func (d *UnixPluginDecoder) Close() error {
	return nil
}

func (d *UnixPluginDecoder) Assets() (map[string][]byte, error) {
	return d.PluginDecoderHelper.Assets(d, "/")
}

func (d *UnixPluginDecoder) CheckAssetsValid() error {
	return nil
}

func (d *UnixPluginDecoder) Checksum() (string, error) {
	return "", nil
}

func (d *UnixPluginDecoder) Manifest() (plugin_entities.PluginDeclaration, error) {
	return plugin_entities.PluginDeclaration{}, nil
}

func (d *UnixPluginDecoder) UniqueIdentity() (plugin_entities.PluginUniqueIdentifier, error) {
	return plugin_entities.PluginUniqueIdentifier(""), nil
}

func (d *UnixPluginDecoder) AvailableI18nReadme() (map[string]string, error) {
	return d.PluginDecoderHelper.AvailableI18nReadme(d, "/")
}

type WindowsPluginDecoder struct {
	UnixPluginDecoder
}

func (d *WindowsPluginDecoder) ReadDir(dirname string) ([]string, error) {
	return []string{
		"_assets\\test.txt",
		"_assets\\test2.txt",
	}, nil
}

func (d *WindowsPluginDecoder) Assets() (map[string][]byte, error) {
	return d.PluginDecoderHelper.Assets(d, "\\")
}

func TestRemapAssets(t *testing.T) {
	decoder := UnixPluginDecoder{}
	remappedAssets, err := decoder.Assets()
	if err != nil {
		t.Fatalf("Failed to remap assets: %v", err)
	}
	assert.Equal(t, remappedAssets["test.txt"], []byte("test"))
	assert.Equal(t, remappedAssets["test2.txt"], []byte("test"))

	decoder1 := WindowsPluginDecoder{}
	remappedAssets, err = decoder1.Assets()
	if err != nil {
		t.Fatalf("Failed to remap assets: %v", err)
	}
	assert.Equal(t, remappedAssets["test.txt"], []byte("test"))
	assert.Equal(t, remappedAssets["test2.txt"], []byte("test"))
}

func TestAvailableI18nReadme(t *testing.T) {
	decoder := UnixPluginDecoder{}
	readmes, err := decoder.AvailableI18nReadme()
	if err != nil {
		t.Fatalf("Failed to get available i18n readme: %v", err)
	}
	assert.Equal(t, readmes["en_US"], "test")
	assert.Equal(t, readmes["zh_Hans"], "test")
}
