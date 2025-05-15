package aliyun

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	dify_oss "github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type AliyunOSSStorage struct {
	client *oss.Client
	bucket *oss.Bucket
	path   string
}

func NewAliyunOSSStorage(
	region string,
	endpoint string,
	accessKeyID string,
	accessKeySecret string,
	authVersion string,
	path string,
	bucketName string,
) (*AliyunOSSStorage, error) {
	// options
	var options []oss.ClientOption

	// set region (required for v4)
	if region != "" {
		options = append(options, oss.Region(region))
	}

	// set auth-version
	if authVersion == "v1" {
		options = append(options, oss.AuthVersion(oss.AuthV1))
	} else if authVersion == "v4" {
		options = append(options, oss.AuthVersion(oss.AuthV4))
	} else {
		// default use v4
		options = append(options, oss.AuthVersion(oss.AuthV4))
	}

	// create client
	var client *oss.Client
	var err error

	client, err = oss.New(endpoint, accessKeyID, accessKeySecret, options...)

	if err != nil {
		return nil, fmt.Errorf("failed to create AliyunOSS client: %w", err)
	}

	// get specified bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket %s: %w", bucketName, err)
	}

	// normalize path: remove leading slash, ensure trailing slash
	path = strings.TrimPrefix(path, "/")
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return &AliyunOSSStorage{
		client: client,
		bucket: bucket,
		path:   path,
	}, nil
}

// combine full object path
func (s *AliyunOSSStorage) fullPath(key string) string {
	return path.Join(s.path, key)
}

func (s *AliyunOSSStorage) Save(key string, data []byte) error {
	fullPath := s.fullPath(key)
	return s.bucket.PutObject(fullPath, bytes.NewReader(data))
}

func (s *AliyunOSSStorage) Load(key string) ([]byte, error) {
	fullPath := s.fullPath(key)
	object, err := s.bucket.GetObject(fullPath)
	if err != nil {
		return nil, err
	}
	// Ensure object is closed after reading
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *AliyunOSSStorage) Exists(key string) (bool, error) {
	fullPath := s.fullPath(key)
	return s.bucket.IsObjectExist(fullPath)
}

func (s *AliyunOSSStorage) State(key string) (dify_oss.OSSState, error) {
	fullPath := s.fullPath(key)
	meta, err := s.bucket.GetObjectMeta(fullPath)
	if err != nil {
		return dify_oss.OSSState{}, err
	}

	// Get content length
	size := int64(0)
	contentLength := meta.Get("Content-Length")
	if contentLength != "" {
		_, err := fmt.Sscanf(contentLength, "%d", &size)
		if err != nil {
			// Return zero size if parsing fails
			size = 0
		}
	}

	// Get last modified time
	lastModified := time.Time{}
	lastModifiedStr := meta.Get("Last-Modified")
	if lastModifiedStr != "" {
		lastModified, err = time.Parse(time.RFC1123, lastModifiedStr)
		if err != nil {
			// Return zero time if parsing fails
			lastModified = time.Time{}
		}
	}

	return dify_oss.OSSState{
		Size:         size,
		LastModified: lastModified,
	}, nil
}

func (s *AliyunOSSStorage) List(prefix string) ([]dify_oss.OSSPath, error) {
	// combine given prefix with path
	fullPrefix := s.fullPath(prefix)

	// Ensure the prefix ends with a slash for directories
	if !strings.HasSuffix(fullPrefix, "/") {
		fullPrefix = fullPrefix + "/"
	}

	var keys []dify_oss.OSSPath
	marker := ""
	for {
		lsRes, err := s.bucket.ListObjects(oss.Marker(marker), oss.Prefix(fullPrefix))
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in Aliyun OSS: %w", err)
		}

		for _, object := range lsRes.Objects {
			if object.Key == fullPrefix {
				continue
			}
			// remove path and prefix from full path, only keep relative path
			key := strings.TrimPrefix(object.Key, fullPrefix)
			// Skip empty keys and directories (keys ending with /)
			if key == "" || strings.HasSuffix(key, "/") {
				continue
			}
			keys = append(keys, dify_oss.OSSPath{
				Path:  key,
				IsDir: false,
			})
		}

		// Check if there are more results
		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}

	return keys, nil
}

func (s *AliyunOSSStorage) Delete(key string) error {
	fullPath := s.fullPath(key)
	return s.bucket.DeleteObject(fullPath)
}

func (s *AliyunOSSStorage) Type() string {
	return dify_oss.OSS_TYPE_ALIYUN_OSS
}
