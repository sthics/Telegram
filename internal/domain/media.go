package domain

import "context"

// MediaRepository defines the interface for object storage operations
type MediaRepository interface {
	// GeneratePresignedURL generates a presigned URL for uploading a file
	GeneratePresignedURL(ctx context.Context, objectName string, contentType string, expiry int64) (string, error)
}
