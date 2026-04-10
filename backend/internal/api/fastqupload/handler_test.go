package fastqupload

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/job/stub"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPostFastqMetadata_201(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"S1","r1_uri":"https://example.com/r1.fq.gz","r2_uri":"https://example.com/r2.fq.gz"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	id, ok := got["job_id"].(string)
	if !ok || id == "" {
		t.Fatalf("job_id: %v", got)
	}
}

func TestPostFastqMetadata_503_NoRepository(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), nil)

	body := `{"sample_id":"S1","r1_uri":"https://a/x","r2_uri":"https://a/y"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", w.Code)
	}
}

func TestPostFastqMetadata_400_Validation(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"","r1_uri":"not-a-uri","r2_uri":"https://example.com/ok"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
	var p problem.Problem
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Type != problem.TypeValidationError {
		t.Fatalf("type %q", p.Type)
	}
	if len(p.Errors) < 2 {
		t.Fatalf("errors: %+v", p.Errors)
	}
	// Deterministic order: field name ascending
	for i := 1; i < len(p.Errors); i++ {
		if p.Errors[i].Field < p.Errors[i-1].Field {
			t.Fatalf("errors not sorted: %+v", p.Errors)
		}
	}
}

func TestPostFastqMetadata_400_UnknownField(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"S1","r1_uri":"https://a/x","r2_uri":"https://a/y","extra":1}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestPostFastqMetadata_400_TrailingJSON(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"S1","r1_uri":"https://a/x","r2_uri":"https://a/y"}`
	body += `{"more":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestPostFastqMetadata_S3URI(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"S1","r1_uri":"s3://bucket/path/r1.fq","r2_uri":"s3://bucket/path/r2.fq"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestPostFastqMetadata_S3URI_EmptyKey(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"S1","r1_uri":"s3://bucket/","r2_uri":"http://x/y"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestPostFastqMetadata_OptionalFieldsStored(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	st := stub.New()
	r := gin.New()
	Register(r.Group("/v1/jobs"), st)

	body := `{"sample_id":"S1","r1_uri":"https://a/r1","r2_uri":"https://a/r2","library_id":"lib-1","platform":"ILLUMINA","notes":"qc ok"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d", w.Code)
	}
	var resp CreateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	id, err := uuid.Parse(resp.JobID)
	if err != nil {
		t.Fatal(err)
	}
	j, err := st.GetByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(j.InputRef, []byte(`"library_id"`)) {
		t.Fatalf("input_ref: %s", j.InputRef)
	}
}
