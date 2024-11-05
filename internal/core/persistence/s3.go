package persistence

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

type S3Wrapper struct {
	client *s3.Client
	bucket string
}

func NewS3Wrapper(region string, access_key string, secret_key string, bucket string) (*S3Wrapper, error) {
	c, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			access_key,
			secret_key,
			"",
		)),
	)
	if err != nil {
		log.Panic("Failed to load AWS S3 config: %v", err)
	}

	s3_client := s3.NewFromConfig(c)
	log.Info("AWS S3 config loaded")

	// check
	_, err = s3_client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Panic("Failed to head bucket: %v", err)
	}

	return &S3Wrapper{
		client: s3_client,
		bucket: bucket,
	}, nil
}

func (s *S3Wrapper) Save(tenant_id string, plugin_checksum string, key string, data []byte) error {
	// save to s3
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Wrapper) Load(tenant_id string, plugin_checksum string, key string) ([]byte, error) {
	// load from s3
	resp, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func (s *S3Wrapper) Delete(tenant_id string, plugin_checksum string, key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Wrapper) StateSize(tenant_id string, plugin_checksum string, key string) (int64, error) {
	// get object size
	resp, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, err
	}

	if resp.ContentLength == nil {
		return 0, fmt.Errorf("content length not found")
	}

	return *resp.ContentLength, nil
}
