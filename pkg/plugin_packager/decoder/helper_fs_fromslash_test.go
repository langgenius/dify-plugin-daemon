package decoder

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFSDecoder_FromSlashBoundary(t *testing.T) {
	root := t.TempDir()
	// Create nested file using OS-native separators
	p := filepath.Join(root, "dir", "sub", "file.txt")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	dec := &FSPluginDecoder{root: root}
	if err := dec.Open(); err != nil {
		t.Fatalf("init open: %v", err)
	}

	// Use forward-slash logical path; decoder should convert via FromSlash at boundary
	b, err := dec.ReadFile("dir/sub/file.txt")
	if err != nil {
		t.Fatalf("ReadFile with forward slashes: %v", err)
	}
	if string(b) != "ok" {
		t.Fatalf("unexpected content: %q", string(b))
	}

	// Stat also accepts forward slashes
	if _, err := dec.Stat("dir/sub/file.txt"); err != nil {
		t.Fatalf("Stat with forward slashes: %v", err)
	}

	// FileReader also accepts forward slashes
	r, err := dec.FileReader("dir/sub/file.txt")
	if err != nil {
		t.Fatalf("FileReader with forward slashes: %v", err)
	}
	defer r.Close()
	data, _ := io.ReadAll(r)
	if string(data) != "ok" {
		t.Fatalf("unexpected reader content: %q", string(data))
	}

	// Negative: traversal should fail at boundary
	if _, err := dec.ReadFile("../file.txt"); err == nil {
		t.Fatalf("expected error for traversal read, got nil")
	}
}
