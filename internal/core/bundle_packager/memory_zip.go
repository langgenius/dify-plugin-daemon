package bundle_packager

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type MemoryZipBundlePackager struct {
	GenericBundlePackager

	zipReader *zip.Reader
}

func NewMemoryZipBundlePackager(zipFile []byte) (*MemoryZipBundlePackager, error) {
	// try read manifest file
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		return nil, err
	}

	manifestFile, err := zipReader.Open("manifest.yaml")
	if err != nil {
		return nil, err
	}
	defer manifestFile.Close()

	manifestBytes, err := io.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	bundle, err := parser.UnmarshalYamlBytes[bundle_entities.Bundle](manifestBytes)
	if err != nil {
		return nil, err
	}

	packager := &MemoryZipBundlePackager{
		GenericBundlePackager: *NewGenericBundlePackager(&bundle),
		zipReader:             zipReader,
	}

	// walk through the zip file and load the assets
	for _, file := range zipReader.File {
		// if file starts with "_assets/"
		if strings.HasPrefix(file.Name, "_assets/") {
			// load the asset
			asset, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer asset.Close()

			assetBytes, err := io.ReadAll(asset)
			if err != nil {
				return nil, err
			}

			// trim the prefix "_assets/"
			assetName := strings.TrimPrefix(file.Name, "_assets/")

			packager.assets[assetName] = bytes.NewBuffer(assetBytes)
		}
	}

	return packager, nil
}

func (p *MemoryZipBundlePackager) Save() error {
	return nil
}
