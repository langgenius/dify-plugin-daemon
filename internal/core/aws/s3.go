package aws

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
)

var (
	s3Client *s3.Client
	s3Bucket *string
)

func InitS3(app *app.Config) {
	// Check if required AWS S3 configuration is provided
	if app.AWSS3Region == nil || app.AWSS3AccessKey == nil || app.AWSS3SecretKey == nil || app.AWSS3Bucket == nil {
		log.Panic("AWSS3Region, AWSS3AccessKey, AWSS3SecretKey, and AWSS3Bucket must be set")
	}

	// Load AWS configuration with provided credentials
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(*app.AWSS3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			*app.AWSS3AccessKey,
			*app.AWSS3SecretKey,
			"",
		)),
	)

	// Handle error if AWS config loading fails
	if err != nil {
		log.Panic("Failed to load AWS S3 config: %v", err)
	}

	log.Info("AWS S3 config loaded")

	// Create S3 client
	s3Client = s3.NewFromConfig(cfg)

	// Store S3 bucket name
	s3Bucket = app.AWSS3Bucket

	log.Info("AWS S3 client initialized successfully")
}

func StreamUploadToS3(ctx context.Context, key string, reader io.Reader) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: s3Bucket,
		Key:    &key,
		Body:   reader,
	})

	return err
}

func StreamDownloadFromS3(ctx context.Context, key string) (io.ReadCloser, error) {
	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: s3Bucket,
		Key:    &key,
	})

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func DeleteFromS3(ctx context.Context, key string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: s3Bucket,
		Key:    &key,
	})

	return err
}

func ListFromS3(ctx context.Context, prefix string) ([]string, error) {
	resp, err := s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: s3Bucket,
		Prefix: &prefix,
	})

	if err != nil {
		return nil, err
	}

	return mapping.MapArray(resp.Contents, func(obj types.Object) string {
		if obj.Key != nil {
			return *obj.Key
		}
		return ""
	}), nil
}

func HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	return s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: s3Bucket,
		Key:    &key,
	})
}
