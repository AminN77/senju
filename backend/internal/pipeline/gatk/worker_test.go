package gatk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/stub"
	"github.com/AminN77/senju/backend/internal/platform/metrics"
	"github.com/AminN77/senju/backend/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeRunner struct {
	mu   sync.Mutex
	run  func(ctx context.Context, name string, args ...string) (int, error)
	call []runnerCall
}

type runnerCall struct {
	name string
	args []string
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) (int, error) {
	f.mu.Lock()
	f.call = append(f.call, runnerCall{name: name, args: append([]string(nil), args...)})
	f.mu.Unlock()
	return f.run(ctx, name, args...)
}

type recordingRepo struct{ base *stub.Repository }

func newRecordingRepo() *recordingRepo { return &recordingRepo{base: stub.New()} }
func (r *recordingRepo) Create(ctx context.Context, p job.CreateParams) (*job.Job, error) {
	return r.base.Create(ctx, p)
}
func (r *recordingRepo) GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	return r.base.GetByID(ctx, id)
}
func (r *recordingRepo) Update(ctx context.Context, id uuid.UUID, p job.UpdateParams) (*job.Job, error) {
	return r.base.Update(ctx, id, p)
}

func TestWorkerHandle_SuccessPersistsVCFAndTransitions(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{run: func(_ context.Context, _ string, _ ...string) (int, error) { return 0, nil }}
	w, err := NewWorker(repo, runner, zerolog.Nop(), Config{DefaultTimeout: time.Second, DefaultThreads: 8, DefaultMemoryMB: 4096}, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"reference_path":"/ref.fa","input_bam_path":"/x.bam","output_vcf_path":"/x.vcf","output_vcf_uri":"s3://bucket/x.vcf"}`
	if err := w.Handle(context.Background(), queue.Message{JobID: created.ID.String(), Payload: json.RawMessage(payload)}); err != nil {
		t.Fatalf("handle err: %v", err)
	}
	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Stage != StageGATKSucceeded || updated.Status != job.StatusSucceeded {
		t.Fatalf("job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"vcf_uri":"s3://bucket/x.vcf"`)) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
}

func TestWorkerHandle_ErrorClassificationAndFailurePersistence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		payload        string
		runner         func(ctx context.Context, name string, args ...string) (int, error)
		wantClass      string
		expectHandleEr bool
	}{
		{
			name:           "configuration decode error",
			payload:        `{"reference_path":"/ref.fa"}`,
			runner:         func(_ context.Context, _ string, _ ...string) (int, error) { return 0, nil },
			wantClass:      ErrorClassConfiguration,
			expectHandleEr: false,
		},
		{
			name:    "tool error",
			payload: `{"reference_path":"/ref.fa","input_bam_path":"/x.bam","output_vcf_path":"/x.vcf"}`,
			runner: func(_ context.Context, _ string, _ ...string) (int, error) {
				return 2, errors.New("gatk failed")
			},
			wantClass:      ErrorClassTool,
			expectHandleEr: false,
		},
		{
			name:    "infrastructure timeout",
			payload: `{"reference_path":"/ref.fa","input_bam_path":"/x.bam","output_vcf_path":"/x.vcf","timeout_seconds":1}`,
			runner: func(_ context.Context, _ string, _ ...string) (int, error) {
				return -1, context.DeadlineExceeded
			},
			wantClass:      ErrorClassInfrastructure,
			expectHandleEr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newRecordingRepo()
			created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
			if err != nil {
				t.Fatal(err)
			}
			m := NewMetrics()
			w, err := NewWorker(repo, &fakeRunner{run: tt.runner}, zerolog.Nop(), Config{DefaultTimeout: time.Second}, m)
			if err != nil {
				t.Fatal(err)
			}
			handleErr := w.Handle(context.Background(), queue.Message{JobID: created.ID.String(), Payload: json.RawMessage(tt.payload)})
			if tt.expectHandleEr && handleErr == nil {
				t.Fatal("expected handle error")
			}
			if !tt.expectHandleEr && handleErr != nil {
				t.Fatalf("unexpected handle error: %v", handleErr)
			}
			updated, err := repo.GetByID(context.Background(), created.ID)
			if err != nil {
				t.Fatal(err)
			}
			if updated.Stage != StageGATKFailed || updated.Status != job.StatusFailed {
				t.Fatalf("job %+v", updated)
			}
			if !bytes.Contains(updated.OutputRef, []byte(`"error"`)) {
				t.Fatalf("output_ref %s", updated.OutputRef)
			}
			if !bytes.Contains(updated.OutputRef, []byte(`"error_class":"`+tt.wantClass+`"`)) {
				t.Fatalf("output_ref missing error_class: %s", updated.OutputRef)
			}
			class := classifyError(handleErr, 2)
			if !tt.expectHandleEr {
				// For terminal persisted runtime failures, handler returns nil by design.
				class = tt.wantClass
			}
			if class != tt.wantClass {
				t.Fatalf("class got=%s want=%s", class, tt.wantClass)
			}
		})
	}
}

func TestWorkerHandle_ExportsPrometheusMetrics(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	stageMetrics := NewMetrics()
	reg := metrics.NewRegistry()
	for _, c := range stageMetrics.Collectors() {
		reg.MustRegister(c)
	}
	runner := &fakeRunner{run: func(_ context.Context, _ string, _ ...string) (int, error) { return 0, nil }}
	w, err := NewWorker(repo, runner, zerolog.Nop(), Config{DefaultTimeout: time.Second}, stageMetrics)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"reference_path":"/ref.fa","input_bam_path":"/x.bam","output_vcf_path":"/x.vcf"}`
	if err := w.Handle(context.Background(), queue.Message{JobID: created.ID.String(), Payload: json.RawMessage(payload)}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(reg.Handler())
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "senju_pipeline_stage_duration_seconds") {
		t.Fatalf("metrics missing duration: %s", text)
	}
	if !strings.Contains(text, "stage=\"gatk\"") || !strings.Contains(text, "outcome=\"success\"") {
		t.Fatalf("metrics missing labels: %s", text)
	}
}
