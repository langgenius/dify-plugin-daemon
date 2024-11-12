package packager

import (
	"archive/zip"
	"bytes"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
)

type Packager struct {
	decoder  decoder.PluginDecoder
	manifest string // manifest file path
}

func NewPackager(decoder decoder.PluginDecoder) *Packager {
	return &Packager{
		decoder:  decoder,
		manifest: "manifest.yaml",
	}
}

func (p *Packager) Pack() ([]byte, error) {
	err := p.Validate()
	if err != nil {
		return nil, err
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	err = p.decoder.Walk(func(filename, dir string) error {
		fullPath := filepath.Join(dir, filename)
		file, err := p.decoder.ReadFile(fullPath)
		if err != nil {
			return err
		}

		zipFile, err := zipWriter.Create(fullPath)
		if err != nil {
			return err
		}

		_, err = zipFile.Write(file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return zipBuffer.Bytes(), nil
}
