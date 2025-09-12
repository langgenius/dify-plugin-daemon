package plugin

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/packager"
)

func PackagePlugin(inputPath string, outputPath string, maxSizeBytes int64) {
	decoder, err := decoder.NewFSPluginDecoder(inputPath)
	if err != nil {
		log.Error("failed to create plugin decoder , plugin path: %s, error: %v", inputPath, err)
		os.Exit(1)
		return
	}

	packager := packager.NewPackager(decoder)
	zipFile, err := packager.Pack(maxSizeBytes)

	if err != nil {
		log.Error("failed to package plugin: %v", err)
		os.Exit(1)
		return
	}

	err = os.WriteFile(outputPath, zipFile, 0644)
	if err != nil {
		log.Error("failed to write package file %v", err)
		os.Exit(1)
		return
	}

	log.Info("plugin packaged successfully, output path: %s", outputPath)
}
