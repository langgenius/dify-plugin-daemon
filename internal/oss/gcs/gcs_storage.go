package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCSStorage struct {
	bucket *storage.BucketHandle
}

func NewGCSStorage(ctx context.Context, bucketName string, opts ...option.ClientOption) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create GCS client: %w", err)
	}

	bucket := client.Bucket(bucketName)
	// check if the bucket exists
	_, err = bucket.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return &GCSStorage{
		bucket: bucket,
	}, nil
}

func (s *GCSStorage) Type() string {
	return oss.OSS_TYPE_GCS
}

func (s *GCSStorage) Save(key string, data []byte) error {
	ctx := context.TODO()
	obj := s.bucket.Object(key)
	w := obj.NewWriter(ctx)
	defer func() {
		if err := w.Close(); err != nil {
			log.Error("failed to close GCS object writer: %v", err)
		}
	}()

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write data to GCS object %s/%s: %w", s.bucket.BucketName(), key, err)
	}
	return nil
}

func (s *GCSStorage) Load(key string) ([]byte, error) {
	ctx := context.TODO()
	obj := s.bucket.Object(key)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GCS object reader %s/%s: %w", s.bucket.BucketName(), key, err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data from GCS object %s/%s: %w", s.bucket.BucketName(), key, err)
	}

	return data, nil
}

func (s *GCSStorage) Exists(key string) (bool, error) {
	ctx := context.TODO()
	obj := s.bucket.Object(key)
	_, err := obj.Attrs(ctx)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, storage.ErrObjectNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("check existence of GCS object %s/%s: %w", s.bucket.BucketName(), key, err)
}

func (s *GCSStorage) State(key string) (oss.OSSState, error) {
	ctx := context.TODO()
	obj := s.bucket.Object(key)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return oss.OSSState{}, fmt.Errorf("get attributes of GCS object %s/%s: %w", s.bucket.BucketName(), key, err)
	}

	state := oss.OSSState{
		Size:         attrs.Size,
		LastModified: attrs.Updated,
	}
	return state, nil
}

func (s *GCSStorage) List(prefix string) ([]oss.OSSPath, error) {
	ctx := context.TODO()
	paths := make([]oss.OSSPath, 0)
	// NOTE: Query prefix must be empty when listing from the root
	if prefix == "/" {
		prefix = ""
	}
	query := &storage.Query{Prefix: prefix}

	it := s.bucket.Objects(ctx, query)
	for {
		fmt.Println("iterating over GCS objects with prefix:", prefix)
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list GCS objects with prefix %s: %w", prefix, err)
		}

		// Skip if it's the prefix itself
		if attrs.Name == prefix {
			continue
		}

		// remove prefix and leading slash
		key := strings.TrimPrefix(attrs.Name, prefix)
		key = strings.TrimPrefix(key, "/")

		paths = append(paths, oss.OSSPath{
			Path:  key,
			IsDir: false,
		})
	}

	return paths, nil
}

func (s *GCSStorage) Delete(key string) error {
	ctx := context.TODO()
	err := s.bucket.Object(key).Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete GCS object %s/%s: %w", s.bucket.BucketName(), key, err)
	}
	return nil
}
