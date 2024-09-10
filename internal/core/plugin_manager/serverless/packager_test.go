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
			origin_file, err := test_plugin.Open(path)
			if err != nil {
				return err
			}
			defer origin_file.Close()

			content, err := io.ReadAll(origin_file)
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

	gzip_reader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gzip_reader.Close()

	// Create a new tar reader
	tar_reader := tar.NewReader(gzip_reader)

	dockerfile_found := false
	requirements_found := false
	main_py_found := false

	// Iterate through the files in the tar.gz archive
	for {
		header, err := tar_reader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			t.Fatal(err)
		}

		switch header.Name {
		case "Dockerfile":
			dockerfile_found = true
		case "requirements.txt":
			requirements_found = true
		case "main.py":
			main_py_found = true
		}
	}

	// Check if all required files are present
	if !dockerfile_found {
		t.Error("Dockerfile not found in the packed archive")
	}
	if !requirements_found {
		t.Error("requirements.txt not found in the packed archive")
	}
	if !main_py_found {
		t.Error("main.py not found in the packed archive")
	}
}
