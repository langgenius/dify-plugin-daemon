package media_transport

import (
	"path"

	"github.com/langgenius/dify-cloud-kit/oss"
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
	filePath := path.Join(m.packagePath, name)

	return m.oss.Save(filePath, file)
}

func (m *PackageBucket) Get(name string) ([]byte, error) {
	return m.oss.Load(path.Join(m.packagePath, name))
}

func (m *PackageBucket) Delete(name string) error {
	// delete from storage
	return m.oss.Delete(path.Join(m.packagePath, name))
}
