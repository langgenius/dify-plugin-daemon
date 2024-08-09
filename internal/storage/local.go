package storage

import (
	"io"
	"os"
)

type Local struct{}

func (l *Local) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (l *Local) ReadStream(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (l *Local) Write(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func (l *Local) WriteStream(path string, data io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, data)
	return err
}

func (l *Local) List(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	file_infos := make([]FileInfo, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		file_infos[i] = info
	}
	return file_infos, nil
}

func (l *Local) Stat(path string) (FileInfo, error) {
	return os.Stat(path)
}

func (l *Local) Delete(path string) error {
	return os.Remove(path)
}

func (l *Local) Mkdir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (l *Local) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (l *Local) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
