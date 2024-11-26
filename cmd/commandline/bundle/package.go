package bundle

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func PackageBundle(bundlePath string, outputPath string) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("Failed to load bundle packager: %v", err)
		os.Exit(1)
		return
	}

	zipFile, err := packager.Export()
	if err != nil {
		log.Error("Failed to export bundle: %v", err)
		os.Exit(1)
		return
	}

	if err := os.WriteFile(outputPath, zipFile, 0644); err != nil {
		log.Error("Failed to write zip file: %v", err)
		os.Exit(1)
		return
	}

	log.Info("Successfully packaged bundle")
}
