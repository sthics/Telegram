package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/ambarg/mini-telegram/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Repository struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func New(ctx context.Context, cfg *config.Config) (*Repository, error) {
	// Load AWS config
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(cfg.ObjectStoreRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.ObjectStoreAccessKey,
			cfg.ObjectStoreSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	// Create main S3 client with internal endpoint for server-side operations
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.ObjectStoreEndpoint)
		o.UsePathStyle = true // Required for MinIO
	})

	// Create separate presign client using PUBLIC endpoint for browser-accessible URLs
	// This ensures presigned URLs are valid when accessed from the browser
	publicEndpoint := cfg.ObjectStorePublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.ObjectStoreEndpoint // Fallback to internal endpoint
	}
	presignClient := s3.NewPresignClient(s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(publicEndpoint)
		o.UsePathStyle = true
	}))

	return &Repository{
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.ObjectStoreBucket,
	}, nil
}

func (r *Repository) GeneratePresignedURL(ctx context.Context, objectName string, contentType string, expirySeconds int64) (string, error) {
	req, err := r.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(objectName),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirySeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}

	return req.URL, nil
}


