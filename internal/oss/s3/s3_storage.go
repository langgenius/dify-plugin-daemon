package s3

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type AWSS3Storage struct {
	bucket string
	client *s3.Client
}

func NewAWSS3Storage(ak string, sk string, region string, bucket string) (oss.OSS, error) {
	var cfg aws.Config
	var err error

	if ak == "" && sk == "" {
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(region),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				ak,
				sk,
				"",
			)),
		)
	}
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	// check bucket
	_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		_, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return nil, err
		}
	}

	return &AWSS3Storage{bucket: bucket, client: client}, nil
}

func (s *AWSS3Storage) Save(key string, data []byte) error {
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (s *AWSS3Storage) Load(key string) ([]byte, error) {
	resp, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func (s *AWSS3Storage) Exists(key string) (bool, error) {
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err == nil, nil
}

func (s *AWSS3Storage) Delete(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *AWSS3Storage) List(prefix string) ([]oss.OSSPath, error) {
	// append a slash to the prefix if it doesn't end with one
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	var keys []oss.OSSPath
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			// remove prefix
			key := strings.TrimPrefix(*obj.Key, prefix)
			// remove leading slash
			key = strings.TrimPrefix(key, "/")
			keys = append(keys, oss.OSSPath{
				Path:  key,
				IsDir: false,
			})
		}
	}

	return keys, nil
}

func (s *AWSS3Storage) State(key string) (oss.OSSState, error) {
	resp, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return oss.OSSState{}, err
	}

	if resp.ContentLength == nil {
		resp.ContentLength = parser.ToPtr[int64](0)
	}
	if resp.LastModified == nil {
		resp.LastModified = parser.ToPtr(time.Time{})
	}

	return oss.OSSState{
		Size:         *resp.ContentLength,
		LastModified: *resp.LastModified,
	}, nil
}
