package packager

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

const generatedRequirementsFilename = "requirements.txt"

type Packager struct {
	decoder  decoder.PluginDecoder
	manifest string // manifest file path
}

type packageFile struct {
	Path    string
	Content []byte
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

	filesToPackage, err := p.collectFiles()
	if err != nil {
		return nil, err
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	totalSize := int64(0)
	var files []FileInfoWithPath

	for _, file := range filesToPackage {
		totalSize, files, err = writePackagedFile(zipWriter, file.Path, file.Content, totalSize, maxSize, files)
		if err != nil {
			return nil, err
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return zipBuffer.Bytes(), nil
}

func (p *Packager) collectFiles() ([]packageFile, error) {
	files := make([]packageFile, 0)
	existingPaths := make(map[string]struct{})

	err := p.decoder.Walk(func(filename, dir string) error {
		fullPath := filepath.Join(dir, filename)
		content, err := p.decoder.ReadFile(fullPath)
		if err != nil {
			return err
		}

		files = append(files, packageFile{
			Path:    fullPath,
			Content: content,
		})
		existingPaths[fullPath] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	generatedFiles, err := p.generatedFiles()
	if err != nil {
		return nil, err
	}

	for _, file := range generatedFiles {
		if _, exists := existingPaths[file.Path]; exists {
			continue
		}
		files = append(files, file)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, nil
}

func (p *Packager) generatedFiles() ([]packageFile, error) {
	manifest, err := p.fetchManifest()
	if err != nil {
		return nil, err
	}

	if manifest.Meta.Runner.Language != constants.Python {
		return nil, nil
	}

	if _, err := p.decoder.Stat(generatedRequirementsFilename); err == nil {
		return []packageFile{}, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if _, err := p.decoder.Stat("pyproject.toml"); err != nil {
		if os.IsNotExist(err) {
			return []packageFile{}, nil
		}
		return nil, err
	}

	if _, err := p.decoder.Stat("uv.lock"); err != nil {
		if os.IsNotExist(err) {
			return []packageFile{}, nil
		}
		return nil, err
	}

	fsDecoder, ok := p.decoder.(*decoder.FSPluginDecoder)
	if !ok {
		return []packageFile{}, nil
	}

	requirements, err := exportRequirementsFromUv(fsDecoder.Root())
	if err != nil {
		return nil, err
	}

	return []packageFile{
		{
			Path:    generatedRequirementsFilename,
			Content: requirements,
		},
	}, nil
}

func exportRequirementsFromUv(root string) ([]byte, error) {
	uvPath := os.Getenv("UV_PATH")
	if uvPath == "" {
		var err error
		uvPath, err = exec.LookPath("uv")
		if err != nil {
			return nil, fmt.Errorf("failed to find uv executable for exporting requirements.txt: %w", err)
		}
	}

	args := []string{
		"export",
		"--frozen",
		"--format", "requirements.txt",
		"--no-group", "dev",
		"--no-emit-project",
		"--no-hashes",
		"--no-header",
		"--no-annotate",
	}

	cmd := exec.Command(uvPath, args...)
	cmd.Dir = root

	output, err := cmd.Output()
	if err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
			return nil, fmt.Errorf("uv export failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("uv export failed: %w", err)
	}

	return output, nil
}

func writePackagedFile(
	zipWriter *zip.Writer,
	fullPath string,
	file []byte,
	totalSize int64,
	maxSize int64,
	files []FileInfoWithPath,
) (int64, []FileInfoWithPath, error) {
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
		return totalSize, files, errors.New(errMsg)
	}

	fullPath = strings.ReplaceAll(fullPath, "\\", "/")

	zipFile, err := zipWriter.Create(fullPath)
	if err != nil {
		return totalSize, files, err
	}

	_, err = zipFile.Write(file)
	if err != nil {
		return totalSize, files, err
	}

	return totalSize, files, nil
}

type FileInfoWithPath struct {
	Path string
	Size int64
}
