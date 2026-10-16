package decoder

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes a file ensuring parent dirs exist.
func writeFile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func minimalManifest(modelProviderPath string) []byte {
	return []byte("" +
		"version: \"0.0.1\"\n" +
		"type: plugin\n" +
		"author: test\n" +
		"name: demo\n" +
		"label:\n  en_US: demo\n" +
		"description:\n  en_US: demo\n" +
		"icon: icon.svg\n" +
		"resource:\n  memory: 134217728\n" +
		"plugins:\n  models:\n    - '" + modelProviderPath + "'\n" +
		"meta:\n  version: \"0.0.1\"\n  arch: [amd64]\n  runner:\n    language: python\n    version: \"3.10\"\n    entrypoint: run.sh\n")
}

func minimalProviderYAML(useWindowsPaths bool) []byte {
	pattern := "models/llm/*.yaml"
	pos := "positions/llm_position.yaml"
	if useWindowsPaths {
		pattern = "models\\llm\\*.yaml"
		pos = "positions\\llm_position.yaml"
	}
	return []byte("" +
		"provider: demo\n" +
		"label:\n  en_US: demo\n" +
		"supported_model_types:\n  - llm\n" +
		"configurate_methods:\n  - predefined-model\n" +
		"models:\n  llm:\n    position: '" + pos + "'\n    predefined:\n      - '" + pattern + "'\n")
}

func minimalModel(id string) []byte {
	return []byte("" +
		"model: " + id + "\n" +
		"label:\n  en_US: " + id + "\n" +
		"model_type: llm\n")
}

func TestManifestModelDiscovery_FSDecoder_UnixAndWindowsPatterns(t *testing.T) {
	root := t.TempDir()
	// Common files
	writeFile(t, root, "icon.svg", []byte("x"))

	// UNIX-style provider
	writeFile(t, root, "manifest.yaml", minimalManifest("provider_unix.yaml"))
	writeFile(t, root, "provider_unix.yaml", minimalProviderYAML(false))
	writeFile(t, root, filepath.Join("models", "llm", "a.yaml"), minimalModel("a"))
	writeFile(t, root, filepath.Join("models", "llm", "b.yaml"), minimalModel("b"))
	writeFile(t, root, filepath.Join("positions", "llm_position.yaml"), []byte("- a\n- b\n"))

	dec, err := NewFSPluginDecoder(root)
	if err != nil {
		t.Fatalf("FS decoder (unix) init: %v", err)
	}
	m, err := dec.Manifest()
	if err != nil {
		t.Fatalf("FS decoder (unix) manifest: %v", err)
	}
	if m.Model == nil || len(m.Model.Models) != 2 {
		t.Fatalf("expected 2 models for unix patterns, got %d", len(m.Model.Models))
	}

	// WINDOWS-style provider (manifest uses backslash in plugins.models entry)
	writeFile(t, root, "manifest.yaml", minimalManifest("provider\\win.yaml"))
	writeFile(t, root, filepath.Join("provider", "win.yaml"), minimalProviderYAML(true))

	dec2, err := NewFSPluginDecoder(root)
	if err != nil {
		t.Fatalf("FS decoder (win) init: %v", err)
	}
	m2, err := dec2.Manifest()
	if err != nil {
		t.Fatalf("FS decoder (win) manifest: %v", err)
	}
	if m2.Model == nil || len(m2.Model.Models) != 2 {
		t.Fatalf("expected 2 models for windows patterns, got %d", len(m2.Model.Models))
	}
}

func TestManifestModelDiscovery_ZipDecoder_WindowsPatterns(t *testing.T) {
	// Build an in-memory zip with forward-slash filenames, but provider uses windows backslashes
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	add := func(name string, data []byte) {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	add("icon.svg", []byte("x"))
	add("manifest.yaml", minimalManifest("provider\\win.yaml"))
	add("provider/win.yaml", minimalProviderYAML(true)) // windows-style paths in provider
	add("models/llm/a.yaml", minimalModel("a"))
	add("models/llm/b.yaml", minimalModel("b"))
	add("positions/llm_position.yaml", []byte("- a\n- b\n"))
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}

	dec, err := NewZipPluginDecoder(buf.Bytes())
	if err != nil {
		t.Fatalf("Zip decoder init: %v", err)
	}
	m, err := dec.Manifest()
	if err != nil {
		t.Fatalf("Zip decoder manifest: %v", err)
	}
	if m.Model == nil || len(m.Model.Models) != 2 {
		t.Fatalf("expected 2 models for windows patterns in zip, got %d", len(m.Model.Models))
	}
}
