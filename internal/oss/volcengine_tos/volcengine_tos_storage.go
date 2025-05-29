package volcengine_tos

import (
	"bytes"
	"context"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"io"
	"strings"
)

type VolcengineTOSStorage struct {
	bucket string
	client *tos.ClientV2
}

func NewVolcengineTOSStorage(endpoint, region, accessKey, secretKey, bucket string) (oss.OSS, error) {
	client, err := tos.NewClientV2(endpoint,
		tos.WithRegion(region),
		tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)),
	)
	if err != nil {
		return nil, err
	}
	return &VolcengineTOSStorage{
		bucket: bucket,
		client: client,
	}, nil
}

func (s *VolcengineTOSStorage) Save(key string, data []byte) error {
	_, err := s.client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: s.bucket,
			Key:    key,
		},
		Content: bytes.NewReader(data),
	})
	return err
}

func (s *VolcengineTOSStorage) Load(key string) ([]byte, error) {
	resp, err := s.client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: s.bucket,
		Key:    key,
	})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Content)
}

func (s *VolcengineTOSStorage) Exists(key string) (bool, error) {
	_, err := s.client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{
		Bucket: s.bucket,
		Key:    key,
	})
	if err != nil {
		if tosErr, ok := err.(*tos.TosServerError); ok && tosErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *VolcengineTOSStorage) State(key string) (oss.OSSState, error) {
	resp, err := s.client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{
		Bucket: s.bucket,
		Key:    key,
	})
	if err != nil {
		return oss.OSSState{}, err
	}
	return oss.OSSState{
		Size:         resp.ContentLength,
		LastModified: resp.LastModified,
	}, nil
}

func (s *VolcengineTOSStorage) List(prefix string) ([]oss.OSSPath, error) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	var result []oss.OSSPath
	truncated := true
	continuationToken := ""
	for truncated {

		resp, err := s.client.ListObjectsType2(context.Background(), &tos.ListObjectsType2Input{
			Bucket:            s.bucket,
			Prefix:            prefix,
			MaxKeys:           1,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range resp.Contents {
			// remove prefix
			key := strings.TrimPrefix(obj.Key, prefix)
			// remove leading slash
			key = strings.TrimPrefix(key, "/")
			if key == "" {
				continue
			}
			result = append(result, oss.OSSPath{
				Path:  key,
				IsDir: false,
			})
		}
		if !resp.IsTruncated {
			break
		}
		truncated = resp.IsTruncated
		continuationToken = resp.NextContinuationToken
	}
	return result, nil
}

func (s *VolcengineTOSStorage) Delete(key string) error {
	_, err := s.client.DeleteObjectV2(context.Background(), &tos.DeleteObjectV2Input{
		Bucket: s.bucket,
		Key:    key,
	})
	return err
}

func (s *VolcengineTOSStorage) Type() string {
	return oss.OSS_TYPE_VOLCENGINE_TOS
}
