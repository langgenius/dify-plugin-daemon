package alioss

import (
	"bytes"
	"fmt"
	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"io"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"
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

	size, _ := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	lastModified, _ := time.Parse(time.RFC1123, header.Get("Last-Modified"))

	return oss.OSSState{
		Size:         size,
		LastModified: lastModified,
	}, nil
}

func (s *AliOSSStorage) List(prefix string) ([]oss.OSSPath, error) {
	fullPrefix := wrapperFolderFilename(s.folder, prefix)
	fullPrefix = strings.TrimSuffix(fullPrefix, "/") + "/"

	var paths []oss.OSSPath
	marker := ""
	maxKeys := 100
	dirSet := make(map[string]struct{})

	for {
		options := []aliyunoss.Option{
			aliyunoss.Prefix(fullPrefix),
			aliyunoss.Marker(marker),
			aliyunoss.MaxKeys(maxKeys),
		}

		lor, err := s.bucket.ListObjects(options...)
		if err != nil {
			return nil, fmt.Errorf("OSS list objects failed: %v", err)
		}

		for _, obj := range lor.Objects {
			if strings.HasSuffix(obj.Key, "/") {
				continue
			}

			relativePath := strings.TrimPrefix(obj.Key, fullPrefix)

			segments := strings.Split(relativePath, "/")

			currentPath := ""
			for i := 0; i < len(segments); i++ {
				if i == len(segments)-1 {
					paths = append(paths, oss.OSSPath{
						Path:  relativePath,
						IsDir: false,
					})
				} else {
					currentPath = path.Join(currentPath, segments[i])
					if _, exists := dirSet[currentPath]; !exists {
						paths = append(paths, oss.OSSPath{
							Path:  currentPath,
							IsDir: true,
						})
						dirSet[currentPath] = struct{}{}
					}
				}
			}
		}

		if !lor.IsTruncated {
			break
		}
		marker = lor.NextMarker
	}

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})

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
