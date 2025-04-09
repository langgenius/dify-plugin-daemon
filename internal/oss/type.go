package oss

import "time"

// OSS supports different types of object storage services
// such as local file system, AWS S3, and Tencent COS.
// The interface defines methods for saving, loading, checking existence,
const (
	OSS_TYPE_LOCAL       = "local"
	OSS_TYPE_S3          = "aws_s3"
	OSS_TYPE_TENCENT_COS = "tencent_cos"
	OSS_TYPE_AZURE_BLOB  = "azure_blob"
)

type OSSState struct {
	Size         int64
	LastModified time.Time
}

type OSSPath struct {
	Path  string
	IsDir bool
}

type OSS interface {
	// Save saves data into path key
	Save(key string, data []byte) error
	// Load loads data from path key
	Load(key string) ([]byte, error)
	// Exists checks if the data exists in the path key
	Exists(key string) (bool, error)
	// State gets the state of the data in the path key
	State(key string) (OSSState, error)
	// List lists all the data with the given prefix, and all the paths are absolute paths
	List(prefix string) ([]OSSPath, error)
	// Delete deletes the data in the path key
	Delete(key string) error
	// Type returns the type of the storage
	// For example: local, aws_s3, tencent_cos
	Type() string
}
