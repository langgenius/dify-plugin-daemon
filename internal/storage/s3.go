package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/aws"
)

type S3 struct{}

func (s *S3) Read(path string) ([]byte, error) {
	reader, err := s.ReadStream(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func (s *S3) ReadStream(path string) (io.ReadCloser, error) {
	return aws.StreamDownloadFromS3(context.Background(), path)
}

func (s *S3) Write(path string, data []byte) error {
	return aws.StreamUploadToS3(context.Background(), path, io.NopCloser(bytes.NewReader(data)))
}

func (s *S3) WriteStream(path string, data io.Reader) error {
	return aws.StreamUploadToS3(context.Background(), path, data)
}

func (s *S3) List(path string) ([]FileInfo, error) {
	keys, err := aws.ListFromS3(context.Background(), path)
	if err != nil {
		return nil, err
	}

	file_infos := make([]FileInfo, len(keys))
	for i, key := range keys {
		head, err := aws.HeadObject(context.Background(), key)
		if err != nil {
			return nil, err
		}
		is_dir := strings.HasSuffix(key, "/")
		file_infos[i] = &s3FileInfo{
			name:    strings.TrimSuffix(key, "/"),
			size:    *head.ContentLength,
			modTime: *head.LastModified,
			isDir:   is_dir,
		}
	}
	return file_infos, nil
}

func (s *S3) Stat(path string) (FileInfo, error) {
	head, err := aws.HeadObject(context.Background(), path)
	if err != nil {
		return nil, err
	}
	return &s3FileInfo{
		name:    path,
		size:    *head.ContentLength,
		modTime: *head.LastModified,
	}, nil
}

func (s *S3) Delete(path string) error {
	return aws.DeleteFromS3(context.Background(), path)
}

func (s *S3) Mkdir(path string, perm os.FileMode) error {
	// S3 doesn't have directories, so this is a no-op
	return nil
}

func (s *S3) Rename(oldpath, newpath string) error {
	// S3 doesn't support rename directly, so we need to copy and delete
	reader, err := s.ReadStream(oldpath)
	if err != nil {
		return err
	}
	defer reader.Close()

	err = aws.StreamUploadToS3(context.Background(), newpath, reader)
	if err != nil {
		return err
	}

	return s.Delete(oldpath)
}

func (s *S3) Exists(path string) (bool, error) {
	_, err := aws.HeadObject(context.Background(), path)
	if err != nil {
		// TODO: Check if error is specifically "not found" error
		return false, nil
	}
	return true, nil
}

type s3FileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (fi *s3FileInfo) Name() string       { return fi.name }
func (fi *s3FileInfo) Size() int64        { return fi.size }
func (fi *s3FileInfo) Mode() os.FileMode  { return 0 }
func (fi *s3FileInfo) ModTime() time.Time { return fi.modTime }
func (fi *s3FileInfo) IsDir() bool        { return fi.isDir }
func (fi *s3FileInfo) Sys() interface{}   { return nil }
