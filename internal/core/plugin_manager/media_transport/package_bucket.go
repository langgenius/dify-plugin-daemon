package media_transport

import (
	"path"
	"runtime"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type PackageBucket struct {
	oss         oss.OSS
	packagePath string
}

func NewPackageBucket(oss oss.OSS, package_path string) *PackageBucket {
	return &PackageBucket{oss: oss, packagePath: package_path}
}

// Save saves a file to the package bucket
func (m *PackageBucket) Save(name string, file []byte) error {
	if runtime.GOOS == "windows" {
		name = strings.ReplaceAll(name, ":", "$")
	}
	filePath := path.Join(m.packagePath, name)

	return m.oss.Save(filePath, file)
}

func (m *PackageBucket) Get(name string) ([]byte, error) {
	if runtime.GOOS == "windows" {
		name = strings.ReplaceAll(name, ":", "$")
	}
	filePath := path.Join(m.packagePath, name)
	return m.oss.Load(filePath)
}

func (m *PackageBucket) Delete(name string) error {
	// delete from storage
	if runtime.GOOS == "windows" {
		name = strings.ReplaceAll(name, ":", "$")
	}
	filePath := path.Join(m.packagePath, name)
	return m.oss.Delete(filePath)
}
