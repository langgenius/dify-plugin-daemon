package tencent

import (
	"bytes"
	"context"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

type TencentCOS struct {
	root   string
	client *cos.Client
}

// NewTencentCOS Create and return an initialized TencentCOS instance.
func NewTencentCOS(bucketURL, secretID, secretKey, root string) (*TencentCOS, error) {
	u, err := url.Parse(bucketURL)
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
	return &TencentCOS{root: root, client: client}, nil
}

func (t *TencentCOS) Save(key string, data []byte) error {
	path := filepath.Join(t.root, key)
	_, err := t.client.Object.Put(context.Background(), path, bytes.NewReader(data), nil)
	return err
}

func (t *TencentCOS) Load(key string) ([]byte, error) {
	path := filepath.Join(t.root, key)
	response, err := t.client.Object.Get(context.Background(), path, nil)
	if err != nil {
		return nil, err
	}
	bs, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (t *TencentCOS) Exists(key string) (bool, error) {
	path := filepath.Join(t.root, key)
	_, err := t.client.Object.Head(context.Background(), path, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (t *TencentCOS) State(key string) (oss.OSSState, error) {
	path := filepath.Join(t.root, key)
	response, err := t.client.Object.Head(context.Background(), path, nil)
	if err != nil {
		return oss.OSSState{}, err
	}

	size := response.Header.Get("Content-Length")
	lastModified := response.Header.Get("Last-Modified")

	sizeInt64, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return oss.OSSState{}, err
	}

	lastModifiedTime, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		return oss.OSSState{}, err
	}

	return oss.OSSState{
		Size:         sizeInt64,
		LastModified: lastModifiedTime,
	}, nil
}

func (t *TencentCOS) List(prefix string) ([]oss.OSSPath, error) {
	path := filepath.Join(t.root, prefix)
	opt := &cos.BucketGetOptions{
		Prefix: path,
	}
	resp, _, err := t.client.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	var paths []oss.OSSPath
	for _, content := range resp.Contents {
		paths = append(paths, oss.OSSPath{
			Path:  content.Key,
			IsDir: false, // COS does not explicitly indicate directories in this context
		})
	}

	return paths, nil
}

func (t *TencentCOS) Delete(key string) error {
	path := filepath.Join(t.root, key)
	_, err := t.client.Object.Delete(context.Background(), path)
	return err
}
