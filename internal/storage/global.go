package storage

import (
	"io"
	"os"
)

var (
	globalStorage FSOperator = &Local{}
)

func Read(path string) ([]byte, error) {
	return globalStorage.Read(path)
}

func ReadStream(path string) (io.ReadCloser, error) {
	return globalStorage.ReadStream(path)
}

func Write(path string, data []byte) error {
	return globalStorage.Write(path, data)
}

func WriteStream(path string, data io.Reader) error {
	return globalStorage.WriteStream(path, data)
}

func List(path string) ([]FileInfo, error) {
	return globalStorage.List(path)
}

func Stat(path string) (FileInfo, error) {
	return globalStorage.Stat(path)
}

func Delete(path string) error {
	return globalStorage.Delete(path)
}

func Mkdir(path string, perm os.FileMode) error {
	return globalStorage.Mkdir(path, perm)
}

func Rename(oldpath, newpath string) error {
	return globalStorage.Rename(oldpath, newpath)
}

func Exists(path string) (bool, error) {
	return globalStorage.Exists(path)
}

func SetGlobalStorage(storage FSOperator) {
	globalStorage = storage
}
