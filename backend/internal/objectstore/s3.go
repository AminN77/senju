package objectstore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Options configures an AWS SDK v2 S3 client for S3-compatible endpoints (e.g. MinIO).
type S3Options struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

// S3Store implements [Service] using the AWS SDK for Go v2 S3 client.
type S3Store struct {
	bucket        string
	client        *s3.Client
	presignClient *s3.PresignClient
	partSizeBytes int64
	presignTTL    time.Duration
}

// NewS3 builds an [S3Store] or returns an error if options are invalid.
func NewS3(opts S3Options) (*S3Store, error) {
	if strings.TrimSpace(opts.Endpoint) == "" || strings.TrimSpace(opts.Bucket) == "" ||
		strings.TrimSpace(opts.AccessKey) == "" || strings.TrimSpace(opts.SecretKey) == "" {
		return nil, errors.New("objectstore: incomplete S3 options")
	}
	region := strings.TrimSpace(opts.Region)
	if region == "" {
		region = "us-east-1"
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(opts.AccessKey),
			strings.TrimSpace(opts.SecretKey),
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("objectstore: aws config: %w", err)
	}

	endpoint := strings.TrimRight(strings.TrimSpace(opts.Endpoint), "/")
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = opts.UsePathStyle
	})

	return &S3Store{
		bucket:        strings.TrimSpace(opts.Bucket),
		client:        client,
		presignClient: s3.NewPresignClient(client),
		partSizeBytes: DefaultPartSizeBytes,
		presignTTL:    PresignTTL,
	}, nil
}

// EnsureBucket creates the configured bucket if it does not already exist (idempotent).
// Intended for local development and integration tests against MinIO.
func (s *S3Store) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}
	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		// Another client may have created the bucket, or MinIO may use a different conflict code.
		if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") ||
			strings.Contains(err.Error(), "BucketAlreadyExists") {
			return nil
		}
		return fmt.Errorf("objectstore: CreateBucket: %w", err)
	}
	return nil
}

// CreateMultipartUpload starts a multipart upload for the given object key.
func (s *S3Store) CreateMultipartUpload(ctx context.Context, objectKey, contentType string) (*MultipartCreateResult, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return nil, errors.New("objectstore: empty object key")
	}
	ct := strings.TrimSpace(contentType)
	if ct == "" {
		ct = "application/octet-stream"
	}

	out, err := s.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(ct),
	})
	if err != nil {
		return nil, fmt.Errorf("objectstore: CreateMultipartUpload: %w", err)
	}
	if out.UploadId == nil || *out.UploadId == "" {
		return nil, errors.New("objectstore: empty upload ID")
	}

	return &MultipartCreateResult{
		Bucket:        s.bucket,
		ObjectKey:     key,
		UploadID:      *out.UploadId,
		PartSizeBytes: s.partSizeBytes,
	}, nil
}

// PresignUploadPart returns a presigned PUT URL for one part.
func (s *S3Store) PresignUploadPart(ctx context.Context, objectKey, uploadID string, partNumber int32) (*PresignedPartResult, error) {
	key := strings.TrimSpace(objectKey)
	uid := strings.TrimSpace(uploadID)
	if key == "" || uid == "" {
		return nil, errors.New("objectstore: object key and upload ID required")
	}
	if partNumber < 1 || partNumber > 10000 {
		return nil, errors.New("objectstore: part number out of range")
	}

	out, err := s.presignClient.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uid),
		PartNumber: aws.Int32(partNumber),
	}, func(o *s3.PresignOptions) {
		o.Expires = s.presignTTL
	})
	if err != nil {
		return nil, fmt.Errorf("objectstore: PresignUploadPart: %w", err)
	}
	if out.URL == "" {
		return nil, errors.New("objectstore: empty presigned URL")
	}

	expires := time.Now().UTC().Add(s.presignTTL)
	return &PresignedPartResult{URL: out.URL, ExpiresAt: expires}, nil
}

// CompleteMultipartUpload completes the multipart upload and returns the composite ETag and size from S3.
func (s *S3Store) CompleteMultipartUpload(ctx context.Context, objectKey, uploadID string, parts []PartETag) (*CompleteResult, error) {
	key := strings.TrimSpace(objectKey)
	uid := strings.TrimSpace(uploadID)
	if key == "" || uid == "" {
		return nil, errors.New("objectstore: object key and upload ID required")
	}
	if len(parts) == 0 {
		return nil, errors.New("objectstore: no parts")
	}

	sorted := append([]PartETag(nil), parts...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PartNumber < sorted[j].PartNumber
	})

	completed := make([]types.CompletedPart, 0, len(sorted))
	for _, p := range sorted {
		if p.PartNumber < 1 || p.PartNumber > 10000 {
			return nil, errors.New("objectstore: invalid part number in completion list")
		}
		etag := strings.TrimSpace(p.ETag)
		if etag == "" {
			return nil, errors.New("objectstore: empty part ETag")
		}
		completed = append(completed, types.CompletedPart{
			ETag:       aws.String(etag),
			PartNumber: aws.Int32(p.PartNumber),
		})
	}

	_, err := s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uid),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completed,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("objectstore: CompleteMultipartUpload: %w", err)
	}

	head, err := s.HeadObject(ctx, key)
	if err != nil {
		return nil, err
	}
	return &CompleteResult{ETag: head.ETag, SizeBytes: head.SizeBytes}, nil
}

// HeadObject returns object metadata.
func (s *S3Store) HeadObject(ctx context.Context, objectKey string) (*HeadResult, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return nil, errors.New("objectstore: empty object key")
	}
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("objectstore: HeadObject: %w", err)
	}
	size := int64(0)
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	etag := ""
	if out.ETag != nil {
		etag = *out.ETag
	}
	return &HeadResult{SizeBytes: size, ETag: etag}, nil
}
