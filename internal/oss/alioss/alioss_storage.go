package alioss

import (
	"bytes"
	"fmt"
	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"io"
	"log"
	"path"
	"strconv"
	"time"
)

type AliOSSStorage struct {
	client *aliyunoss.Client
	bucket *aliyunoss.Bucket
	folder string
}

func NewOSSStorage(ak string, sk string, endpoint string, bkt string, folder string) (oss.OSS, error) {
	client, err := aliyunoss.New(endpoint, ak, sk)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %v", err)
	}

	bucket, err := client.Bucket(bkt)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %v", err)
	}

	return &AliOSSStorage{
		client: client,
		bucket: bucket,
		folder: folder,
	}, nil
}

func (s *AliOSSStorage) Save(key string, data []byte) error {
	return s.bucket.PutObject(wrapperFolderFilename(s.folder, key), bytes.NewReader(data))
}

func (s *AliOSSStorage) Load(key string) ([]byte, error) {
	body, err := s.bucket.GetObject(wrapperFolderFilename(s.folder, key))
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Printf("failed to close object body for key %q: %v", wrapperFolderFilename(s.folder, key), err)
		}
	}(body)
	return io.ReadAll(body)
}

func (s *AliOSSStorage) Exists(key string) (bool, error) {
	return s.bucket.IsObjectExist(wrapperFolderFilename(s.folder, key))
}

func (s *AliOSSStorage) State(key string) (oss.OSSState, error) {
	header, err := s.bucket.GetObjectMeta(wrapperFolderFilename(s.folder, key))
	if err != nil {
		return oss.OSSState{}, fmt.Errorf("failed to get object metadata: %w", err)
	}

	// 解析文件大小和最后修改时间
	size, _ := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	lastModified, _ := time.Parse(time.RFC1123, header.Get("Last-Modified"))

	return oss.OSSState{
		Size:         size,
		LastModified: lastModified,
	}, nil
}

func (s *AliOSSStorage) List(prefix string) ([]oss.OSSPath, error) {
	var paths []oss.OSSPath

	marker := aliyunoss.Marker("")
	for {
		lor, err := s.bucket.ListObjects(aliyunoss.Prefix(wrapperFolderFilename(s.folder, prefix)), aliyunoss.Delimiter("/"), marker)
		if err != nil {
			return nil, fmt.Errorf("list objects failed: %w", err)
		}

		// 处理文件
		for _, obj := range lor.Objects {
			paths = append(paths, oss.OSSPath{
				Path:  obj.Key,
				IsDir: false,
			})
		}

		// 处理目录
		for _, prefix := range lor.CommonPrefixes {
			paths = append(paths, oss.OSSPath{
				Path:  prefix,
				IsDir: true,
			})
		}

		marker = aliyunoss.Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	return paths, nil
}

func (s *AliOSSStorage) Delete(key string) error {
	return s.bucket.DeleteObject(wrapperFolderFilename(s.folder, key))
}

func wrapperFolderFilename(folder string, filename string) string {
	if folder != "" {
		return path.Join(folder, filename)
	}
	return filename
}
