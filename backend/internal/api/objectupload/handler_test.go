package objectupload

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AminN77/senju/backend/internal/objectstore"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type fakeStore struct {
	createFn   func(ctx context.Context, objectKey, contentType string) (*objectstore.MultipartCreateResult, error)
	presignFn  func(ctx context.Context, objectKey, uploadID string, partNumber int32) (*objectstore.PresignedPartResult, error)
	completeFn func(ctx context.Context, objectKey, uploadID string, parts []objectstore.PartETag) (*objectstore.CompleteResult, error)
	headFn     func(ctx context.Context, objectKey string) (*objectstore.HeadResult, error)
}

func (f *fakeStore) CreateMultipartUpload(ctx context.Context, objectKey, contentType string) (*objectstore.MultipartCreateResult, error) {
	if f.createFn != nil {
		return f.createFn(ctx, objectKey, contentType)
	}
	return nil, nil
}

func (f *fakeStore) PresignUploadPart(ctx context.Context, objectKey, uploadID string, partNumber int32) (*objectstore.PresignedPartResult, error) {
	if f.presignFn != nil {
		return f.presignFn(ctx, objectKey, uploadID, partNumber)
	}
	return nil, nil
}

func (f *fakeStore) CompleteMultipartUpload(ctx context.Context, objectKey, uploadID string, parts []objectstore.PartETag) (*objectstore.CompleteResult, error) {
	if f.completeFn != nil {
		return f.completeFn(ctx, objectKey, uploadID, parts)
	}
	return nil, nil
}

func (f *fakeStore) HeadObject(ctx context.Context, objectKey string) (*objectstore.HeadResult, error) {
	if f.headFn != nil {
		return f.headFn(ctx, objectKey)
	}
	return nil, nil
}

func TestRegister_DisabledReturns503(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/objects"), nil, zerolog.Nop())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart", strings.NewReader(`{"content_type":"application/octet-stream"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", w.Code)
	}
}

func TestMultipartInit_Validation(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/objects"), &fakeStore{
		createFn: func(_ context.Context, _, _ string) (*objectstore.MultipartCreateResult, error) {
			return &objectstore.MultipartCreateResult{
				Bucket: "b", ObjectKey: "k", UploadID: "u", PartSizeBytes: 64 << 20,
			}, nil
		},
	}, zerolog.Nop())

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"missing content_type", `{}`, http.StatusBadRequest},
		{"ok", `{"content_type":"application/gzip"}`, http.StatusCreated},
		{"unknown field", `{"content_type":"application/gzip","extra":1}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status %d body %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestPresignPart_Validation(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/objects"), &fakeStore{
		presignFn: func(_ context.Context, _, _ string, _ int32) (*objectstore.PresignedPartResult, error) {
			return &objectstore.PresignedPartResult{URL: "https://example/p", ExpiresAt: time.Now().Add(time.Minute)}, nil
		},
	}, zerolog.Nop())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart/up1/parts", strings.NewReader(`{"object_key":"","part_number":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestComplete_AuditLogFields(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	buf := bytes.Buffer{}
	log := zerolog.New(&buf)
	r := gin.New()
	Register(r.Group("/v1/objects"), &fakeStore{
		completeFn: func(_ context.Context, _, _ string, _ []objectstore.PartETag) (*objectstore.CompleteResult, error) {
			return &objectstore.CompleteResult{ETag: `"abc"`, SizeBytes: 5}, nil
		},
	}, log)

	w := httptest.NewRecorder()
	body := `{"object_key":"uploads/x/y","parts":[{"part_number":1,"etag":"\"etag1\""}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart/up99/complete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	out := buf.String()
	if !strings.Contains(out, "object_multipart_complete") || !strings.Contains(out, "checksum_status") {
		t.Fatalf("expected audit fields in log: %q", out)
	}
}

func TestBuildObjectKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		hint *string
	}{
		{"no hint", nil},
		{"empty hint", ptr("  ")},
		{"hint", ptr("file-name_R1.fq.gz")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k := buildObjectKey(tt.hint)
			if !strings.HasPrefix(k, "uploads/") {
				t.Fatalf("got %q", k)
			}
		})
	}
}

func TestSanitizeFilenameHint(t *testing.T) {
	t.Parallel()
	if got := sanitizeFilenameHint("a b<c>"); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func ptr(s string) *string { return &s }

func BenchmarkMultipartInit_Throughput(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	Register(r.Group("/v1/objects"), &fakeStore{
		createFn: func(_ context.Context, objectKey, _ string) (*objectstore.MultipartCreateResult, error) {
			return &objectstore.MultipartCreateResult{
				Bucket: "b", ObjectKey: objectKey, UploadID: "u", PartSizeBytes: objectstore.DefaultPartSizeBytes,
			}, nil
		},
	}, zerolog.Nop())
	body := `{"content_type":"application/octet-stream"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			b.Fatalf("status %d", w.Code)
		}
	}
}

func TestDecodeJSON_TrailingRejected(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/objects"), &fakeStore{}, zerolog.Nop())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart", strings.NewReader(`{"content_type":"x"} {}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestMultipartInitResponseShape(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/objects"), &fakeStore{
		createFn: func(_ context.Context, objectKey, ct string) (*objectstore.MultipartCreateResult, error) {
			if ct != "application/octet-stream" {
				t.Fatalf("ct %q", ct)
			}
			if objectKey == "" {
				t.Fatal("empty key")
			}
			return &objectstore.MultipartCreateResult{
				Bucket:        "mybucket",
				ObjectKey:     objectKey,
				UploadID:      "uid-1",
				PartSizeBytes: objectstore.DefaultPartSizeBytes,
			}, nil
		},
	}, zerolog.Nop())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/objects/multipart", strings.NewReader(`{"content_type":"application/octet-stream"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var got MultipartInitResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Bucket != "mybucket" || got.UploadID != "uid-1" || got.PartSizeBytes != objectstore.DefaultPartSizeBytes {
		t.Fatalf("%+v", got)
	}
}
