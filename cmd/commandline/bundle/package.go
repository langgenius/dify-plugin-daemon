package bundle

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func PackageBundle(bundlePath string, outputPath string) {
	packager, err := loadBundlePackager(bundlePath)
	if err != nil {
		log.Error("failed to load bundle packager", "error", err)
		os.Exit(1)
		return
	}

	zipFile, err := packager.Export()
	if err != nil {
		log.Error("failed to export bundle", "error", err)
		os.Exit(1)
		return
	}

	if err := os.WriteFile(outputPath, zipFile, 0644); err != nil {
		log.Error("failed to write zip file", "error", err)
		os.Exit(1)
		return
	}

	log.Info("successfully packaged bundle")
}
