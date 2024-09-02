package decoder

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

var (
	ErrNotDir = errors.New("not a directory")
)

type FSPluginDecoder struct {
	PluginDecoder
	PluginDecoderHelper

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

	// read the manifest file
	if _, err := decoder.Manifest(); err != nil {
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
		if info.IsDir() {
			return nil
		}

		// trim the first directory path
		path = strings.TrimPrefix(path, d.root)
		// trim / from the beginning
		path = strings.TrimPrefix(path, "/")

		if path == "" {
			return nil
		}

		if err != nil {
			return err
		}

		return fn(info.Name(), filepath.Dir(path))
	})
}

func (d *FSPluginDecoder) Close() error {
	return nil
}

func (d *FSPluginDecoder) Stat(filename string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(d.root, filename))
}

func (d *FSPluginDecoder) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(d.root, filename))
}

func (d *FSPluginDecoder) FileReader(filename string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(d.root, filename))
}

func (d *FSPluginDecoder) Signature() (string, error) {
	return "", nil
}

func (d *FSPluginDecoder) CreateTime() (int64, error) {
	return 0, nil
}

func (d *FSPluginDecoder) Manifest() (plugin_entities.PluginDeclaration, error) {
	return d.PluginDecoderHelper.Manifest(d)
}
