package decoder

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrNotDir = errors.New("not a directory")
)

type FSPluginDecoder struct {
	PluginDecoder

	// root directory of the plugin
	root string

	fs fs.FS
}

func NewFSPluginDecoder(root string) (*FSPluginDecoder, error) {
	decoder := &FSPluginDecoder{
		root: root,
	}

	err := decoder.Open()
	if err != nil {
		return nil, err
	}

	return decoder, nil
}

func (d *FSPluginDecoder) Open() error {
	d.fs = os.DirFS(d.root)

	// try to stat the root directory
	s, err := os.Stat(d.root)
	if err != nil {
		return err
	}

	if !s.IsDir() {
		return ErrNotDir
	}

	return nil
}

func (d *FSPluginDecoder) Walk(fn func(filename string, dir string) error) error {
	return filepath.Walk(d.root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return fn(info.Name(), filepath.Dir(path))
	})
}

func (d *FSPluginDecoder) Close() error {
	return nil
}

func (d *FSPluginDecoder) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(d.root, filename))
}

func (d *FSPluginDecoder) Signature() (string, error) {
	return "", nil
}

func (d *FSPluginDecoder) CreateTime() (int64, error) {
	return 0, nil
}
