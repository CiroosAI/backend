package utils

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// UploadToS3 uploads a file to S3-compatible object storage using AWS SDK Go v2 (Signature V4)
func UploadToS3(objectName string, file io.Reader, fileSize int64) error {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "forums"
	}

	if endpoint == "" || accessKey == "" || secretKey == "" {
		return fmt.Errorf("S3 config missing in environment variables")
	}

	contentType := mime.TypeByExtension(path.Ext(objectName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Custom resolver for S3-compatible endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: "us-east-1", // default region for S3-compatible
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // for most S3-compatible storage
	})

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:       &bucket,
		Key:          &objectName,
		Body:         file,
		ContentType:  &contentType,
		StorageClass: s3types.StorageClassStandard,
		// keep objects private by not setting ACL; enable server-side encryption
		ServerSideEncryption: s3types.ServerSideEncryptionAes256,
	})
	if err != nil {
		return fmt.Errorf("S3 upload failed: %w", err)
	}
	return nil
}

// GenerateSignedURL returns a presigned GET URL for the given object name and expiry duration
func GenerateSignedURL(objectName string, expirySeconds int64) (string, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "sf-forums"
	}

	if endpoint == "" || accessKey == "" || secretKey == "" {
		return "", fmt.Errorf("S3 config missing in environment variables")
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: "us-east-1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	presigner := s3.NewPresignClient(client)
	in := &s3.GetObjectInput{Bucket: &bucket, Key: &objectName}
	presigned, err := presigner.PresignGetObject(context.TODO(), in, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(expirySeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign s3 url: %w", err)
	}
	return presigned.URL, nil
}

// UploadToS3AndPresign uploads the provided reader to S3 and returns a presigned GET URL
func UploadToS3AndPresign(objectName string, file io.ReadSeeker, fileSize int64, expirySeconds int64) (string, error) {
	// Upload first
	if err := UploadToS3(objectName, file, fileSize); err != nil {
		return "", err
	}
	// Ensure reader is reset before presigning (caller should provide ReadSeeker)
	// Return presigned URL
	url, err := GenerateSignedURL(objectName, expirySeconds)
	if err != nil {
		return "", err
	}
	return url, nil
}
