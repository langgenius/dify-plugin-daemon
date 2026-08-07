package decoder

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func buildTestZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("create entry %q: %v", name, err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("write entry %q: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func TestExtractToRejectsPathTraversal(t *testing.T) {
	dst := t.TempDir()
	outside := filepath.Join(filepath.Dir(dst), "evil-escaped-marker.txt")
	defer os.Remove(outside)

	dec := &ZipPluginDecoder{
		reader: mustZipReader(t, buildTestZip(t, map[string]string{
			"../../evil-escaped-marker.txt": "pwned",
		})),
	}

	err := dec.ExtractTo(dst)
	if err == nil {
		t.Fatal("expected traversal entry to be rejected, got nil error")
	}
	if _, statErr := os.Stat(outside); !os.IsNotExist(statErr) {
		t.Fatalf("traversal entry escaped dst: %s exists", outside)
	}
}

func TestExtractToAllowsNormalEntries(t *testing.T) {
	dst := t.TempDir()
	dec := &ZipPluginDecoder{
		reader: mustZipReader(t, buildTestZip(t, map[string]string{
			"manifest.yaml":  "name: test",
			"sub/dir/ok.txt": "fine",
		})),
	}

	if err := dec.ExtractTo(dst); err != nil {
		t.Fatalf("normal entries rejected: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "sub", "dir", "ok.txt"))
	if err != nil || string(got) != "fine" {
		t.Fatalf("nested entry not extracted correctly: %v %q", err, got)
	}
}

func mustZipReader(t *testing.T, b []byte) *zip.Reader {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	return r
}
