package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	appconfig "github.com/starfrag-lab/retrowin-go/internal/config"
	apperrors "github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
)

// S3Storage implements the Storage interface using AWS S3.
type S3Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

// New creates a new S3 storage instance.
func New(cfg *appconfig.StorageConfig) (storage.Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage config is required")
	}

	// Load AWS configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint if provided
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// For MinIO or local S3-compatible services
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	// Create presign client
	presigner := s3.NewPresignClient(client)

	return &S3Storage{
		client:    client,
		presigner: presigner,
		bucket:    cfg.Bucket,
	}, nil
}

// GetPresignedUploadURL generates a presigned URL for direct client upload.
func (s *S3Storage) GetPresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}
	return req.URL, nil
}

// GetPresignedDownloadURL generates a presigned URL for direct client download.
func (s *S3Storage) GetPresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}
	return req.URL, nil
}

// DeleteObject removes an object from storage.
func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// ObjectExists checks if an object exists in storage.
func (s *S3Storage) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// GetObjectSize returns the size of an object in bytes.
func (s *S3Storage) GetObjectSize(ctx context.Context, key string) (int64, error) {
	resp, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return 0, apperrors.NotFound("object not found")
		}
		return 0, fmt.Errorf("failed to get object info: %w", err)
	}
	if resp.ContentLength == nil {
		return 0, nil
	}
	return *resp.ContentLength, nil
}
