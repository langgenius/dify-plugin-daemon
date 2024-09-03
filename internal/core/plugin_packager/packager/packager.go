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

	zip_buffer := new(bytes.Buffer)
	zip_writer := zip.NewWriter(zip_buffer)

	err = p.decoder.Walk(func(filename, dir string) error {
		full_path := filepath.Join(dir, filename)
		file, err := p.decoder.ReadFile(full_path)
		if err != nil {
			return err
		}

		zip_file, err := zip_writer.Create(full_path)
		if err != nil {
			return err
		}

		_, err = zip_file.Write(file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = zip_writer.Close()
	if err != nil {
		return nil, err
	}

	return zip_buffer.Bytes(), nil
}
