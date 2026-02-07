package system

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilePath(t *testing.T) {
	filepathJoinRes := filepath.Join("foo", "bar.txt")
	pathJoinRs := path.Join("foo", "bar.txt")
	if runtime.GOOS == "windows" {
		assert.Equal(t, "foo\\bar.txt", filepathJoinRes)
	}
	assert.Equal(t, "foo/bar.txt", pathJoinRs)
}

func TestConvertPath(t *testing.T) {
	filepathJoinRes := filepath.Join("foo", "bar.txt")
	res := ConvertPath(filepathJoinRes)
	assert.Equal(t, "foo/bar.txt", res)
}
