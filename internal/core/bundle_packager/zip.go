package bundle_packager

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/bundle_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type ZipBundlePackager struct {
	GenericBundlePackager

	zipReader *zip.Reader
	path      string
}

func NewZipBundlePackager(path string) (BundlePackager, error) {
	zipFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipFileInfo, err := zipFile.Stat()
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(zipFile, zipFileInfo.Size())
	if err != nil {
		return nil, err
	}

	// try read manifest file
	manifestFile, err := zipReader.Open("manifest.json")
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

	packager := &ZipBundlePackager{
		GenericBundlePackager: *NewGenericBundlePackager(&bundle),
		zipReader:             zipReader,
		path:                  path,
	}

	// walk through the zip file and load the assets
	for _, file := range zipReader.File {
		// if file starts with "assets/"
		if strings.HasPrefix(file.Name, "assets/") {
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

			// trim the prefix "assets/"
			assetName := strings.TrimPrefix(file.Name, "assets/")

			packager.assets[assetName] = bytes.NewBuffer(assetBytes)
		}
	}

	return packager, nil
}

func (p *ZipBundlePackager) Save() error {
	// export the bundle to a zip file
	zipBytes, err := p.Export()
	if err != nil {
		return err
	}

	// save the zip file
	err = os.WriteFile(p.path, zipBytes, 0644)
	if err != nil {
		return err
	}

	// reload zip reader
	zipFile, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipFileInfo, err := zipFile.Stat()
	if err != nil {
		return err
	}

	p.zipReader, err = zip.NewReader(zipFile, zipFileInfo.Size())
	if err != nil {
		return err
	}

	return nil
}
