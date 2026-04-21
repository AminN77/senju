package fastqupload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/job"
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

// failingRepo always errors on Create to assert 500 problem+json contract.
type failingRepo struct{}

func (failingRepo) Create(context.Context, job.CreateParams) (*job.Job, error) {
	return nil, errors.New("injected persistence failure")
}

func (failingRepo) GetByID(context.Context, uuid.UUID) (*job.Job, error) {
	return nil, job.ErrNotFound
}

func (failingRepo) Update(context.Context, uuid.UUID, job.UpdateParams) (*job.Job, error) {
	return nil, job.ErrNotFound
}

func TestPostFastqMetadata_500_RepositoryError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), failingRepo{})

	body := `{"sample_id":"S1","r1_uri":"https://a/x","r2_uri":"https://a/y"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var p problem.Problem
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Type != problem.TypeInternalError {
		t.Fatalf("type %q", p.Type)
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

func TestPostFastqMetadata_URIWhitespaceTrimmed(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	body := `{"sample_id":"S1","r1_uri":" s3://bucket/k/r1 ","r2_uri":"s3://bucket/k/r2"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/metadata", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
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

func TestPostFastqValidate_200_Valid(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	st := stub.New()
	created, err := st.Create(context.Background(), job.CreateParams{
		Status: job.StatusPending,
		Stage:  StageFastqUploadQueued,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	Register(r.Group("/v1/jobs"), st)

	body := "@r1\nACGT\n+\nIIII\n"
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/"+created.ID.String()+"/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var got ValidateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !got.Valid || got.Records != 1 {
		t.Fatalf("response %+v", got)
	}
	updated, err := st.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Stage != StageFastqValidated || updated.Status != job.StatusSucceeded {
		t.Fatalf("updated job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"kind":"fastq_validation_v1"`)) {
		t.Fatalf("output_ref: %s", updated.OutputRef)
	}
}

func TestPostFastqValidate_200_Invalid(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	st := stub.New()
	created, err := st.Create(context.Background(), job.CreateParams{
		Status: job.StatusPending,
		Stage:  StageFastqUploadQueued,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	Register(r.Group("/v1/jobs"), st)

	body := "@r1\nACGT\n+\nII\n"
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/"+created.ID.String()+"/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var got ValidateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Valid || got.FailureReason != "quality_length_mismatch" {
		t.Fatalf("response %+v", got)
	}
	updated, err := st.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Stage != StageFastqValidationFailed || updated.Status != job.StatusFailed {
		t.Fatalf("updated job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"failure_reason":"quality_length_mismatch"`)) {
		t.Fatalf("output_ref missing failure_reason: %s", updated.OutputRef)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"failure_record":1`)) {
		t.Fatalf("output_ref missing failure_record: %s", updated.OutputRef)
	}
}

func TestPostFastqValidate_400_BadJobID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1/jobs"), stub.New())

	req := httptest.NewRequest(http.MethodPost, "/v1/jobs/fastq-upload/not-a-uuid/validate", strings.NewReader("@r\nA\n+\n!\n"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}
