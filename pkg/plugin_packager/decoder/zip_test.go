package decoder

import (
	"os"
	"path"
	"testing"
)

func TestExtractFile(t *testing.T) {
	pluginFile, err := os.ReadFile("testdata/github#0.3.2@1cb2f90ea05bbc7987fd712aff0a07594073816269125603dc2fa5b4229eb122")
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	decoder, err := NewZipPluginDecoder(pluginFile)
	if err != nil {
		t.Fatalf("create new zip decoder error: %v", err)
	}
	extractPath := "testdata/cwd"
	err = os.Mkdir(extractPath, 0755)
	if err != nil {
		t.Fatalf("mk dir error: %v", err)
	}
	defer os.RemoveAll(extractPath)

	err = decoder.ExtractTo(extractPath)
	if err != nil {
		t.Fatalf("extract file error: %v", err)
	}
	_, err = os.Stat(path.Join(extractPath, "provider", "github.yaml"))
	if err != nil {
		t.Fatalf("extract file not exists: %v", err)
	}
}
