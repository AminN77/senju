package fastqc

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
	"github.com/AminN77/senju/backend/internal/pipeline/stagemetrics"
	pmetrics "github.com/AminN77/senju/backend/internal/platform/metrics"
	"github.com/AminN77/senju/backend/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeRunner struct {
	run func(ctx context.Context, args ...string) (int, error)
}

func TestWorkerHandle_ExportsPrometheusMetrics(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	stageMetrics := stagemetrics.New()
	reg := pmetrics.NewRegistry()
	for _, c := range stageMetrics.Collectors() {
		reg.MustRegister(c)
	}
	w, err := NewWorker(repo, fakeRunner{
		run: func(_ context.Context, _ ...string) (int, error) { return 0, nil },
	}, zerolog.Nop(), Config{DefaultTimeout: time.Second}, stageMetrics)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"input_path":"/tmp/r1.fastq.gz","output_dir":"/tmp/reports"}`
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
	if !strings.Contains(text, "stage=\"fastqc\"") || !strings.Contains(text, "outcome=\"success\"") {
		t.Fatalf("metrics missing labels: %s", text)
	}
}

func (f fakeRunner) Run(ctx context.Context, args ...string) (int, error) {
	return f.run(ctx, args...)
}

type recordingRepo struct {
	base *stub.Repository

	mu             sync.Mutex
	updateStatuses []job.Status
	updateCtxErrs  []error
}

func newRecordingRepo() *recordingRepo {
	return &recordingRepo{base: stub.New()}
}

func (r *recordingRepo) Create(ctx context.Context, p job.CreateParams) (*job.Job, error) {
	return r.base.Create(ctx, p)
}

func (r *recordingRepo) GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	return r.base.GetByID(ctx, id)
}

func (r *recordingRepo) Update(ctx context.Context, id uuid.UUID, p job.UpdateParams) (*job.Job, error) {
	r.mu.Lock()
	r.updateStatuses = append(r.updateStatuses, p.Status)
	r.updateCtxErrs = append(r.updateCtxErrs, ctx.Err())
	r.mu.Unlock()
	return r.base.Update(ctx, id, p)
}

func TestWorkerHandle_SuccessStoresArtifactsAndLogs(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{
		Status: job.StatusPending,
		Stage:  "queued",
	})
	if err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)
	w, err := NewWorker(repo, fakeRunner{
		run: func(_ context.Context, args ...string) (int, error) {
			if strings.Join(args, " ") != "--outdir /tmp/reports /tmp/r1.fastq.gz" {
				t.Fatalf("args %v", args)
			}
			return 0, nil
		},
	}, logger, Config{DefaultTimeout: time.Second}, nil)
	if err != nil {
		t.Fatal(err)
	}

	payload := `{"input_path":"/tmp/r1.fastq.gz","output_dir":"/tmp/reports","report_html_uri":"s3://bucket/r1_fastqc.html","report_zip_uri":"s3://bucket/r1_fastqc.zip"}`
	err = w.Handle(context.Background(), queue.Message{
		JobID:   created.ID.String(),
		Payload: json.RawMessage(payload),
	})
	if err != nil {
		t.Fatalf("handle err: %v", err)
	}

	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != job.StatusSucceeded || updated.Stage != StageFastQCSucceeded {
		t.Fatalf("job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"html_uri":"s3://bucket/r1_fastqc.html"`)) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
	if !strings.Contains(logBuf.String(), `"job_id":"`+created.ID.String()+`"`) {
		t.Fatalf("log %s", logBuf.String())
	}
	if !strings.Contains(logBuf.String(), `"stage":"fastqc"`) {
		t.Fatalf("log %s", logBuf.String())
	}
	if !strings.Contains(logBuf.String(), `"exit_code":0`) {
		t.Fatalf("log %s", logBuf.String())
	}
}

func TestWorkerHandle_TimeoutAndCancellationEnforced(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{
		Status: job.StatusPending,
		Stage:  "queued",
	})
	if err != nil {
		t.Fatal(err)
	}
	w, err := NewWorker(repo, fakeRunner{
		run: func(ctx context.Context, _ ...string) (int, error) {
			<-ctx.Done()
			return -1, ctx.Err()
		},
	}, zerolog.Nop(), Config{DefaultTimeout: 20 * time.Millisecond}, nil)
	if err != nil {
		t.Fatal(err)
	}

	payload := `{"input_path":"/tmp/r1.fastq.gz","output_dir":"/tmp/reports","timeout_seconds":0}`
	err = w.Handle(context.Background(), queue.Message{
		JobID:   created.ID.String(),
		Payload: json.RawMessage(payload),
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.updateStatuses) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(repo.updateStatuses))
	}
	if repo.updateStatuses[1] != job.StatusFailed {
		t.Fatalf("second update status = %s", repo.updateStatuses[1])
	}
	if repo.updateCtxErrs[1] != nil {
		t.Fatalf("completion update used expired context: %v", repo.updateCtxErrs[1])
	}

	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != job.StatusFailed || updated.Stage != StageFastQCFailed {
		t.Fatalf("job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"exit_code":-1`)) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
}

func TestWorkerHandle_InvalidPayload(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	worker, err := NewWorker(repo, fakeRunner{
		run: func(_ context.Context, _ ...string) (int, error) { return 0, nil },
	}, zerolog.Nop(), Config{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = worker.Handle(context.Background(), queue.Message{
		JobID:   uuid.NewString(),
		Payload: json.RawMessage(`{"output_dir":"/tmp/reports"}`),
	})
	if err == nil || !strings.Contains(err.Error(), "input_path is required") {
		t.Fatalf("unexpected err %v", err)
	}
}
