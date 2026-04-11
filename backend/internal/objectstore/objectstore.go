// Package objectstore defines S3-compatible multipart upload operations for presigned client uploads.
package objectstore

import (
	"context"
	"time"
)

// DefaultPartSizeBytes is the recommended part size for multipart uploads (64 MiB).
// With S3's 10,000 part limit this supports very large objects while staying under minimum part size rules.
const DefaultPartSizeBytes int64 = 64 << 20

// PresignTTL is how long presigned part URLs remain valid.
const PresignTTL = 15 * time.Minute

// MultipartCreateResult is returned after initiating a multipart upload.
type MultipartCreateResult struct {
	Bucket        string
	ObjectKey     string
	UploadID      string
	PartSizeBytes int64
}

// PresignedPartResult is returned for a single part PUT URL.
type PresignedPartResult struct {
	URL       string
	ExpiresAt time.Time
}

// PartETag identifies an uploaded part for completion.
type PartETag struct {
	PartNumber int32
	ETag       string
}

// CompleteResult is returned after completing a multipart upload and verifying the object.
type CompleteResult struct {
	ETag      string
	SizeBytes int64
}

// HeadResult is metadata from HeadObject.
type HeadResult struct {
	SizeBytes int64
	ETag      string
}

// Service abstracts S3 multipart and presign operations.
type Service interface {
	CreateMultipartUpload(ctx context.Context, objectKey, contentType string) (*MultipartCreateResult, error)
	PresignUploadPart(ctx context.Context, objectKey, uploadID string, partNumber int32) (*PresignedPartResult, error)
	CompleteMultipartUpload(ctx context.Context, objectKey, uploadID string, parts []PartETag) (*CompleteResult, error)
	HeadObject(ctx context.Context, objectKey string) (*HeadResult, error)
}
