package media_transport

import (
	"runtime"
	"testing"

	"github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/stretchr/testify/assert"
)

func TestPluginListWindonws(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	config := &app.Config{
		PluginStorageLocalRoot: "testdata",
		PluginInstalledPath:    "plugin",
		PluginStorageType:      "local",
	}
	var storage oss.OSS
	var err error
	storage, err = factory.Load(config.PluginStorageType, oss.OSSArgs{
		Local: &oss.Local{
			Path: config.PluginStorageLocalRoot,
		},
	})
	if err != nil {
		t.Fatal("failed to create storage")
	}
	installedBucket := NewInstalledBucket(storage, config.PluginInstalledPath)
	identifiers, err := installedBucket.List()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(identifiers), 1)
	assert.Equal(t, identifiers[0].String(), "langgenius/github#0.3.2@1cb2f90ea05bbc7987fd712aff0a07594073816269125603dc2fa5b4229eb122")
}
