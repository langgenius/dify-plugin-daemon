package decoder

import (
	"bufio"
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
	// dify_ignores is a map[string][]string, the key is the directory, the value is a list of files to ignore
	difyIgnores := make(map[string][]string)

	return filepath.Walk(d.root, func(path string, info fs.FileInfo, err error) error {
		// trim the first directory path
		path = strings.TrimPrefix(path, d.root)
		// trim / from the beginning
		path = strings.TrimPrefix(path, "/")
		p := filepath.Dir(path)

		if info.IsDir() {
			// try read the .difyignore file if it's the first time to walk this directory
			if _, ok := difyIgnores[p]; !ok {
				difyIgnores[p] = make([]string, 0)
				// read the .difyignore file if it exists
				ignoreFilePath := filepath.Join(d.root, p, ".difyignore")
				if _, err := os.Stat(ignoreFilePath); err == nil {
					ignoreFile, err := os.Open(ignoreFilePath)
					if err != nil {
						return err
					}

					scanner := bufio.NewScanner(ignoreFile)
					for scanner.Scan() {
						line := scanner.Text()
						if strings.HasPrefix(line, "#") {
							continue
						}
						difyIgnores[p] = append(difyIgnores[p], line)
					}

					ignoreFile.Close()
				}
			}

			return nil
		}

		currentIgnoreFiles := difyIgnores[p]
		for _, ignoreFile := range currentIgnoreFiles {
			// skip if match
			matched, err := filepath.Match(ignoreFile, info.Name())
			if err != nil {
				return err
			}
			if matched {
				return nil
			}
		}

		if path == "" {
			return nil
		}

		if err != nil {
			return err
		}

		return fn(info.Name(), p)
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

func (d *FSPluginDecoder) ReadDir(dirname string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(
		filepath.Join(d.root, dirname),
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, err := filepath.Rel(d.root, path)
				if err != nil {
					return err
				}
				files = append(files, relPath)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return files, nil
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

func (d *FSPluginDecoder) Assets() (map[string][]byte, error) {
	return d.PluginDecoderHelper.Assets(d)
}

func (d *FSPluginDecoder) Checksum() (string, error) {
	return d.PluginDecoderHelper.Checksum(d)
}

func (d *FSPluginDecoder) UniqueIdentity() (plugin_entities.PluginUniqueIdentifier, error) {
	return d.PluginDecoderHelper.UniqueIdentity(d)
}

func (d *FSPluginDecoder) CheckAssetsValid() error {
	return d.PluginDecoderHelper.CheckAssetsValid(d)
}
