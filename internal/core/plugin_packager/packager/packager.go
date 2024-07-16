package packager

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path"
)

type Packager struct {
	wp string // working path

	manifest string // manifest file path
}

func NewPackager(plugin_path string) *Packager {
	return &Packager{
		wp:       plugin_path,
		manifest: "manifest.yaml",
	}
}

func (p *Packager) Pack() ([]byte, error) {
	// read manifest
	_, err := p.fetchManifest()
	if err != nil {
		return nil, err
	}

	zip_buffer := new(bytes.Buffer)
	zip_writer := zip.NewWriter(zip_buffer)
	err = fs.WalkDir(os.DirFS(p.wp), ".", func(root_path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		file, err := os.ReadFile(path.Join(p.wp, root_path))
		if err != nil {
			return err
		}

		zip_file, err := zip_writer.Create(root_path)
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
