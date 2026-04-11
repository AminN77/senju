package objectstore

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestS3Store_MultipartRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	if os.Getenv("RUN_MINIO_INTEGRATION") != "1" {
		t.Skip("set RUN_MINIO_INTEGRATION=1 and S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY")
	}
	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET"))
	access := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY"))
	if access == "" {
		access = strings.TrimSpace(os.Getenv("MINIO_ROOT_USER"))
	}
	secret := strings.TrimSpace(os.Getenv("S3_SECRET_KEY"))
	if secret == "" {
		secret = strings.TrimSpace(os.Getenv("MINIO_ROOT_PASSWORD"))
	}
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	if endpoint == "" || bucket == "" || access == "" || secret == "" {
		t.Skip("S3 endpoint, bucket, and credentials required")
	}

	ctx := context.Background()
	store, err := NewS3(S3Options{
		Endpoint:     endpoint,
		Region:       region,
		Bucket:       bucket,
		AccessKey:    access,
		SecretKey:    secret,
		UsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureBucket(ctx); err != nil {
		t.Fatal(err)
	}

	key := "uploads/integration-test/" + uuid.New().String()
	create, err := store.CreateMultipartUpload(ctx, key, "application/octet-stream")
	if err != nil {
		t.Fatal(err)
	}

	pres, err := store.PresignUploadPart(ctx, key, create.UploadID, 1)
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("hello")
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, pres.URL, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.ContentLength = int64(len(payload))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("put part: status %d body %s", resp.StatusCode, b)
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag from part upload")
	}

	res, err := store.CompleteMultipartUpload(ctx, key, create.UploadID, []PartETag{
		{PartNumber: 1, ETag: etag},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.SizeBytes != int64(len(payload)) {
		t.Fatalf("size: got %d want %d", res.SizeBytes, len(payload))
	}
}
