package packager

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
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

	var files []FileInfoWithPath

	err = p.decoder.Walk(func(filename, dir string) error {
		fullPath := filepath.Join(dir, filename)
		file, err := p.decoder.ReadFile(fullPath)
		if err != nil {
			return err
		}
		fileSize := int64(len(file))
		files = append(files, FileInfoWithPath{Path: fullPath, Size: fileSize})
		totalSize += fileSize
		if totalSize > maxSize {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Size > files[j].Size
			})
			fileTop5Info := ""
			top := 5
			if len(files) < 5 {
				top = len(files)
			}
			for i := 0; i < top; i++ {
				fileTop5Info += fmt.Sprintf("%d. name: %s, size: %d bytes\n", i+1, files[i].Path, files[i].Size)
			}
			errMsg := fmt.Sprintf("Plugin package size is too large. Please ensure the uncompressed size is less than %d bytes.\nPackaged file info:\n%s",
				maxSize, fileTop5Info)
			return errors.New(errMsg)
		}

		// ISSUES: Windows path separator is \, but zip requires /, to avoid this we just simply replace all \ with / for now
		// TODO: find a better solution
		fullPath = strings.ReplaceAll(fullPath, "\\", "/")

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

type FileInfoWithPath struct {
	Path string
	Size int64
}
