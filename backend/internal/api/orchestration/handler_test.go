package orchestration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/stub"
	"github.com/gin-gonic/gin"
)

func TestOrchestration_EndToEndSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	repo := stub.New()
	r := gin.New()
	Register(r.Group("/v1/jobs"), repo)

	createBody := `{"sample_id":"S-1","r1_uri":"s3://bucket/r1.fq.gz","r2_uri":"s3://bucket/r2.fq.gz"}`
	createReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/pipeline", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	r.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status %d body %s", createResp.Code, createResp.Body.String())
	}
	var created createResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	runReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+created.JobID+"/run", nil)
	runResp := httptest.NewRecorder()
	r.ServeHTTP(runResp, runReq)
	if runResp.Code != http.StatusOK {
		t.Fatalf("run status %d body %s", runResp.Code, runResp.Body.String())
	}
	var run runResponse
	if err := json.Unmarshal(runResp.Body.Bytes(), &run); err != nil {
		t.Fatal(err)
	}
	if run.Status != job.StatusSucceeded || run.Stage != stagePipelineSucceeded {
		t.Fatalf("run response %+v", run)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+created.JobID+"/status", nil)
	statusResp := httptest.NewRecorder()
	r.ServeHTTP(statusResp, statusReq)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("status code %d body %s", statusResp.Code, statusResp.Body.String())
	}
	var status statusResponse
	if err := json.Unmarshal(statusResp.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Status != job.StatusSucceeded || status.CompletedAt == nil || status.StartedAt == nil {
		t.Fatalf("status %+v", status)
	}

	outputsReq := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+created.JobID+"/outputs", nil)
	outputsResp := httptest.NewRecorder()
	r.ServeHTTP(outputsResp, outputsReq)
	if outputsResp.Code != http.StatusOK {
		t.Fatalf("outputs code %d body %s", outputsResp.Code, outputsResp.Body.String())
	}
	var out struct {
		OutputRef map[string]any `json:"output_ref"`
	}
	if err := json.Unmarshal(outputsResp.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.OutputRef["result"] != "succeeded" {
		t.Fatalf("unexpected output_ref %+v", out.OutputRef)
	}
}

func TestOrchestration_EndToEndControlledFailure(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	repo := stub.New()
	r := gin.New()
	Register(r.Group("/v1/jobs"), repo)

	createBody := `{"sample_id":"S-1","r1_uri":"s3://bucket/r1.fq.gz","r2_uri":"s3://bucket/r2.fq.gz","force_fail":true}`
	createReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/pipeline", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	r.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status %d body %s", createResp.Code, createResp.Body.String())
	}
	var created createResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	runReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+created.JobID+"/run", nil)
	runResp := httptest.NewRecorder()
	r.ServeHTTP(runResp, runReq)
	if runResp.Code != http.StatusOK {
		t.Fatalf("run status %d body %s", runResp.Code, runResp.Body.String())
	}
	var run runResponse
	if err := json.Unmarshal(runResp.Body.Bytes(), &run); err != nil {
		t.Fatal(err)
	}
	if run.Status != job.StatusFailed || run.Stage != stagePipelineFailed {
		t.Fatalf("run response %+v", run)
	}

	outputsReq := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+created.JobID+"/outputs", nil)
	outputsResp := httptest.NewRecorder()
	r.ServeHTTP(outputsResp, outputsReq)
	if outputsResp.Code != http.StatusOK {
		t.Fatalf("outputs code %d body %s", outputsResp.Code, outputsResp.Body.String())
	}
	var out struct {
		OutputRef map[string]any `json:"output_ref"`
	}
	if err := json.Unmarshal(outputsResp.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.OutputRef["result"] != "failed" || out.OutputRef["error"] != "forced_failure" {
		t.Fatalf("unexpected output_ref %+v", out.OutputRef)
	}
}

func TestOrchestration_RunConflict(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	repo := stub.New()
	r := gin.New()
	Register(r.Group("/v1/jobs"), repo)

	createBody := `{"sample_id":"S-1","r1_uri":"s3://bucket/r1.fq.gz","r2_uri":"s3://bucket/r2.fq.gz"}`
	createReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/pipeline", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	r.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create status %d body %s", createResp.Code, createResp.Body.String())
	}
	var created createResponse
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	firstRunReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+created.JobID+"/run", nil)
	firstRunResp := httptest.NewRecorder()
	r.ServeHTTP(firstRunResp, firstRunReq)
	if firstRunResp.Code != http.StatusOK {
		t.Fatalf("first run status %d body %s", firstRunResp.Code, firstRunResp.Body.String())
	}

	secondRunReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+created.JobID+"/run", nil)
	secondRunResp := httptest.NewRecorder()
	r.ServeHTTP(secondRunResp, secondRunReq)
	if secondRunResp.Code != http.StatusConflict {
		t.Fatalf("second run status %d body %s", secondRunResp.Code, secondRunResp.Body.String())
	}
	var p problem.Problem
	if err := json.Unmarshal(secondRunResp.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Type != problem.TypeValidationError {
		t.Fatalf("problem %+v", p)
	}
}
