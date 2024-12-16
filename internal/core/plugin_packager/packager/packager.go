package packager

import (
	"archive/zip"
	"bytes"
	"errors"
	"path/filepath"
	"strconv"

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

func (p *Packager) Pack(maxSize int64) ([]byte, error) {
	err := p.Validate()
	if err != nil {
		return nil, err
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	totalSize := int64(0)

	err = p.decoder.Walk(func(filename, dir string) error {
		fullPath := filepath.Join(dir, filename)
		file, err := p.decoder.ReadFile(fullPath)
		if err != nil {
			return err
		}

		totalSize += int64(len(file))
		if totalSize > maxSize {
			return errors.New("plugin package size is too large, please ensure the uncompressed size is less than " + strconv.FormatInt(maxSize, 10) + " bytes")
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
