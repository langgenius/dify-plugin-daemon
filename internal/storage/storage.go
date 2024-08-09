package storage

import (
	"io"
	"os"
	"time"
)

// FileInfo represents information about a file
type FileInfo interface {
	Name() string
	Size() int64
	Mode() os.FileMode
	ModTime() time.Time
	IsDir() bool
}

// FSOperator defines the interface for basic file system operations
type FSOperator interface {
	// Read operations
	Read(path string) ([]byte, error)
	ReadStream(path string) (io.ReadCloser, error)

	// Write operations
	Write(path string, data []byte) error
	WriteStream(path string, data io.Reader) error

	// List operation
	List(path string) ([]FileInfo, error)

	// Get file info
	Stat(path string) (FileInfo, error)

	// Delete operation
	Delete(path string) error

	// Create directory
	Mkdir(path string, perm os.FileMode) error

	// Rename operation
	Rename(oldpath, newpath string) error

	// Check if file/directory exists
	Exists(path string) (bool, error)
}

// FullFSOperator extends FSOperator with additional operations
type FullFSOperator interface {
	FSOperator

	// Copy operation
	Copy(src, dst string) error

	// Move operation
	Move(src, dst string) error

	// Recursive delete
	DeleteAll(path string) error

	// Create file
	Create(path string) (io.WriteCloser, error)

	// Open file with specific flag and permission
	OpenFile(path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)

	// Get file checksum
	Checksum(path string) (string, error)

	// Watch for file changes
	Watch(path string) (<-chan FileInfo, error)
}
