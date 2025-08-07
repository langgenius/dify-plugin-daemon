package decoder

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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
	// read .difyignore file
	ignorePatterns := []gitignore.Pattern{}
	// Try .difyignore first, fallback to .gitignore if not found
	ignoreBytes, err := d.ReadFile(".difyignore")
	if err != nil {
		ignoreBytes, err = d.ReadFile(".gitignore")
	}
	if err == nil {
		ignoreLines := strings.Split(string(ignoreBytes), "\n")
		for _, line := range ignoreLines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			ignorePatterns = append(ignorePatterns, gitignore.ParsePattern(line, nil))
		}
	}

	return filepath.WalkDir(d.root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// get relative path from root
		relPath, err := filepath.Rel(d.root, path)
		if err != nil {
			return err
		}

		// skip root directory
		if relPath == "." {
			return nil
		}

		// check if path matches any ignore pattern
		pathParts := strings.Split(relPath, string(filepath.Separator))
		for _, pattern := range ignorePatterns {
			if result := pattern.Match(pathParts, info.IsDir()); result == gitignore.Exclude {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// get directory path relative to root
		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = ""
		}

		// skip if the path is a directory
		if info.IsDir() {
			return nil
		}

		return fn(info.Name(), dir)
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

func (d *FSPluginDecoder) Verification() (*Verification, error) {
	return nil, nil
}

func (d *FSPluginDecoder) Manifest() (plugin_entities.PluginDeclaration, error) {
	return d.PluginDecoderHelper.Manifest(d)
}

func (d *FSPluginDecoder) Assets() (map[string][]byte, error) {
	// use filepath.Separator as the separator to make it os-independent
	return d.PluginDecoderHelper.Assets(d, string(filepath.Separator))
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

func (d *FSPluginDecoder) Verified() bool {
	return d.PluginDecoderHelper.verified(d)
}

func (d *FSPluginDecoder) AvailableI18nReadme() (map[string]string, error) {
	return d.PluginDecoderHelper.AvailableI18nReadme(d, string(filepath.Separator))
}
