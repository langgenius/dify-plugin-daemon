package serverless

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
)

//go:embed packager_test_plugin/*
var test_plugin embed.FS

func TestPackager_Pack(t *testing.T) {
	// create a temp dir
	tmpDir, err := os.MkdirTemp("", "test_plugin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// copy the test_plugin to the temp dir
	if err := fs.WalkDir(test_plugin, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// create the dir
			os.MkdirAll(filepath.Join(tmpDir, path), 0755)
		} else {
			// copy the file
			originFile, err := test_plugin.Open(path)
			if err != nil {
				return err
			}
			defer originFile.Close()

			content, err := io.ReadAll(originFile)
			if err != nil {
				return err
			}

			if err := os.WriteFile(filepath.Join(tmpDir, path), content, 0644); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	decoder, err := decoder.NewFSPluginDecoder(path.Join(tmpDir, "packager_test_plugin"))
	if err != nil {
		t.Fatal(err)
	}

	packager := NewPackager(decoder)

	f, err := packager.Pack()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gzipReader.Close()

	// Create a new tar reader
	tarReader := tar.NewReader(gzipReader)

	dockerfileFound := false
	requirementsFound := false
	mainPyFound := false
	jinaYamlFound := false
	// Iterate through the files in the tar.gz archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			t.Fatal(err)
		}

		switch header.Name {
		case "Dockerfile":
			dockerfileFound = true
		case "requirements.txt":
			requirementsFound = true
		case "main.py":
			mainPyFound = true
		case "provider/jina.yaml":
			jinaYamlFound = true
		}
	}

	// Check if all required files are present
	if !dockerfileFound {
		t.Error("Dockerfile not found in the packed archive")
	}
	if !requirementsFound {
		t.Error("requirements.txt not found in the packed archive")
	}
	if !mainPyFound {
		t.Error("main.py not found in the packed archive")
	}
	if !jinaYamlFound {
		t.Error("jina.yaml not found in the packed archive")
	}
}
