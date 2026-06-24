package decoder

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildZipPlugin(t *testing.T, files map[string][]byte) []byte {
	t.Helper()

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	for name, content := range files {
		if len(content) == 0 && strings.HasSuffix(name, "/") {
			_, err := zipWriter.Create(name)
			require.NoError(t, err)
			continue
		}

		writer, err := zipWriter.Create(name)
		require.NoError(t, err)
		_, err = writer.Write(content)
		require.NoError(t, err)
	}
	require.NoError(t, zipWriter.Close())

	return buffer.Bytes()
}

func minimalPluginFiles(t *testing.T) map[string][]byte {
	t.Helper()

	manifest, err := os.ReadFile(filepath.Join("..", "testdata", "manifest.yaml"))
	require.NoError(t, err)
	endpoint, err := os.ReadFile(filepath.Join("..", "testdata", "neko.yaml"))
	require.NoError(t, err)

	return map[string][]byte{
		"manifest.yaml": manifest,
		"neko.yaml":     endpoint,
	}
}

func TestZipPluginDecoderExtractToRejectsParentPath(t *testing.T) {
	files := minimalPluginFiles(t)
	files["../escaped.txt"] = []byte("escaped")

	zipDecoder, err := NewZipPluginDecoder(buildZipPlugin(t, files))
	require.NoError(t, err)

	parent := t.TempDir()
	dst := filepath.Join(parent, "plugin")
	err = zipDecoder.ExtractTo(dst)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errUnsafeZipPath))
	assert.NoFileExists(t, filepath.Join(parent, "escaped.txt"))
	assert.NoDirExists(t, dst)
}

func TestZipPluginDecoderExtractToRejectsBackslashPath(t *testing.T) {
	files := minimalPluginFiles(t)
	files[`..\escaped.txt`] = []byte("escaped")

	zipDecoder, err := NewZipPluginDecoder(buildZipPlugin(t, files))
	require.NoError(t, err)

	parent := t.TempDir()
	dst := filepath.Join(parent, "plugin")
	err = zipDecoder.ExtractTo(dst)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errUnsafeZipPath))
	assert.NoFileExists(t, filepath.Join(parent, "escaped.txt"))
	assert.NoDirExists(t, dst)
}

func TestZipPluginDecoderExtractToAllowsNestedPath(t *testing.T) {
	files := minimalPluginFiles(t)
	files["nested/"] = nil
	files["nested/file.txt"] = []byte("ok")

	zipDecoder, err := NewZipPluginDecoder(buildZipPlugin(t, files))
	require.NoError(t, err)

	dst := filepath.Join(t.TempDir(), "plugin")
	require.NoError(t, zipDecoder.ExtractTo(dst))

	extracted, err := os.ReadFile(filepath.Join(dst, "nested", "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, []byte("ok"), extracted)
}

func TestSafeEntryPathRejectsParentDirectoryEntry(t *testing.T) {
	_, err := safeEntryPath("..")

	require.Error(t, err)
	assert.True(t, errors.Is(err, errUnsafeZipPath))
}

func TestCopyZipFileRejectsContentBeyondDeclaredSize(t *testing.T) {
	var out bytes.Buffer

	err := copyZipFile(&out, strings.NewReader("toolarge"), 3)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds declared uncompressed size")
	assert.Equal(t, "tool", out.String())
}

func TestCopyZipFileAllowsDeclaredSize(t *testing.T) {
	var out bytes.Buffer

	err := copyZipFile(&out, io.NopCloser(strings.NewReader("ok")), 2)

	require.NoError(t, err)
	assert.Equal(t, "ok", out.String())
}

func TestZipPluginDecoderReadDirSkipsDirectoryEntries(t *testing.T) {
	files := minimalPluginFiles(t)
	files["nested/"] = nil
	files["nested/file.txt"] = []byte("ok")

	zipDecoder, err := NewZipPluginDecoder(buildZipPlugin(t, files))
	require.NoError(t, err)

	entries, err := zipDecoder.ReadDir("nested")
	require.NoError(t, err)
	assert.Equal(t, []string{"nested/file.txt"}, entries)
}

func TestZipPluginDecoderWalkSkipsDirectoryEntries(t *testing.T) {
	files := minimalPluginFiles(t)
	files["nested/"] = nil
	files["nested/file.txt"] = []byte("ok")

	zipDecoder, err := NewZipPluginDecoder(buildZipPlugin(t, files))
	require.NoError(t, err)

	visited := make([]string, 0, 2)
	err = zipDecoder.Walk(func(filename, dir string) error {
		visited = append(visited, filepath.Join(dir, filename))
		return nil
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"manifest.yaml", "neko.yaml", filepath.Join("nested", "file.txt")}, visited)
}
