package tencent_cos

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type TencentCOSStorage struct {
	bucket string
	region string
	client *cos.Client
}

func NewTencentCOSStorage(secretID string, secretKey string, region string, bucket string) (oss.OSS, error) {
	u, err := url.Parse("https://" + bucket + ".cos." + region + ".myqcloud.com")
	if err != nil {
		return nil, err
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	_, err = client.Bucket.Head(context.Background())
	if err != nil {
		return nil, err
	}

	return &TencentCOSStorage{
		bucket: bucket,
		region: region,
		client: client,
	}, nil
}

func (s *TencentCOSStorage) Save(key string, data []byte) error {
	_, err := s.client.Object.Put(context.Background(), key, bytes.NewReader(data), nil)
	return err
}

func (s *TencentCOSStorage) Load(key string) ([]byte, error) {
	resp, err := s.client.Object.Get(context.Background(), key, nil)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func (s *TencentCOSStorage) Exists(key string) (bool, error) {
	ok, err := s.client.Object.IsExist(context.Background(), key)
	if err == nil && ok {
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		return false, nil
	}
}

func (s *TencentCOSStorage) Delete(key string) error {
	_, err := s.client.Object.Delete(context.Background(), key)
	return err
}

func (s *TencentCOSStorage) List(prefix string) ([]oss.OSSPath, error) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	var keys []oss.OSSPath
	opt := &cos.BucketGetOptions{
		Prefix:    prefix,
		Delimiter: "/",
	}
	isTruncated := true
	var marker string
	for isTruncated {
		if marker != "" {
			opt.Marker = marker
		}

		result, _, err := s.client.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}

		for _, content := range result.Contents {
			// remove prefix
			key := strings.TrimPrefix(content.Key, prefix)
			// remove leading slash
			key = strings.TrimPrefix(key, "/")
			if key == "" {
				continue
			}
			keys = append(keys, oss.OSSPath{
				Path:  key,
				IsDir: false,
			})
		}

		for _, commonPrefix := range result.CommonPrefixes {
			if commonPrefix == "" {
				continue
			}
			if !strings.HasSuffix(commonPrefix, "/") {
				commonPrefix = commonPrefix + "/"
			}
			keys = append(keys, oss.OSSPath{
				Path:  commonPrefix,
				IsDir: true,
			})

			subKeys, _ := s.List(commonPrefix)
			if len(subKeys) > 0 {
				subPrefix := strings.TrimPrefix(commonPrefix, prefix)
				for i := range subKeys {
					subKeys[i].Path = subPrefix + subKeys[i].Path
				}
				keys = append(keys, subKeys...)
			}

		}

		isTruncated = result.IsTruncated
		marker = result.NextMarker
	}

	return keys, nil
}

func (s *TencentCOSStorage) State(key string) (oss.OSSState, error) {
	resp, err := s.client.Object.Head(context.Background(), key, nil)
	if err != nil {
		return oss.OSSState{}, err
	}

	contentLength := resp.ContentLength

	lastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		lastModified = time.Time{}
	}

	return oss.OSSState{
		Size:         contentLength,
		LastModified: lastModified,
	}, nil
}

func (s *TencentCOSStorage) Type() string {
	return oss.OSS_TYPE_TENCENT_COS
}
