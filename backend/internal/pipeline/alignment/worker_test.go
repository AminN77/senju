package alignment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	mu   sync.Mutex
	run  func(ctx context.Context, name string, args ...string) (int, error)
	call []runnerCall
}

func TestWorkerHandle_ExportsPrometheusMetrics(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	bamPath := filepath.Join(t.TempDir(), "out.bam")
	stageMetrics := stagemetrics.New()
	reg := pmetrics.NewRegistry()
	for _, c := range stageMetrics.Collectors() {
		reg.MustRegister(c)
	}
	runner := &fakeRunner{
		run: func(_ context.Context, name string, _ ...string) (int, error) {
			if name == "samtools" {
				if err := os.WriteFile(bamPath, []byte("bam"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			return 0, nil
		},
	}
	w, err := NewWorker(repo, runner, zerolog.Nop(), Config{DefaultTimeout: time.Second}, stageMetrics)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"reference_path":"/ref.fa","read1_path":"/r1","read2_path":"/r2","output_bam_path":"` + bamPath + `"}`
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
	if !strings.Contains(text, "stage=\"alignment\"") || !strings.Contains(text, "outcome=\"success\"") {
		t.Fatalf("metrics missing labels: %s", text)
	}
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

type recordingRepo struct {
	base *stub.Repository
}

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

func TestWorkerHandle_SuccessPersistsBAMAndChecksum(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{
		Status: job.StatusPending,
		Stage:  "queued",
	})
	if err != nil {
		t.Fatal(err)
	}
	bamPath := filepath.Join(t.TempDir(), "out.bam")

	runner := &fakeRunner{
		run: func(_ context.Context, name string, _ ...string) (int, error) {
			if name == "samtools" {
				if err := os.WriteFile(bamPath, []byte("deterministic-bam"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			return 0, nil
		},
	}
	var logBuf bytes.Buffer
	w, err := NewWorker(repo, runner, zerolog.New(&logBuf), Config{
		DefaultTimeout:  10 * time.Second,
		DefaultThreads:  8,
		DefaultMemoryMB: 4096,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	msgPayload := `{"reference_path":"/ref.fa","read1_path":"/r1.fq.gz","read2_path":"/r2.fq.gz","output_bam_path":"` + bamPath + `","output_bam_uri":"s3://bucket/out.bam"}`
	err = w.Handle(context.Background(), queue.Message{
		JobID:   created.ID.String(),
		Payload: json.RawMessage(msgPayload),
	})
	if err != nil {
		t.Fatalf("handle err: %v", err)
	}

	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Stage != StageAlignmentSucceeded || updated.Status != job.StatusSucceeded {
		t.Fatalf("job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"bam_uri":"s3://bucket/out.bam"`)) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"checksum_sha256"`)) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
	if !strings.Contains(logBuf.String(), `"stage":"alignment"`) || !strings.Contains(logBuf.String(), `"exit_code":0`) {
		t.Fatalf("log %s", logBuf.String())
	}
}

func TestWorkerHandle_ConfigurableLimitsApplied(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	bamPath := filepath.Join(t.TempDir(), "out.bam")
	runner := &fakeRunner{
		run: func(_ context.Context, name string, _ ...string) (int, error) {
			if name == "samtools" {
				if err := os.WriteFile(bamPath, []byte("bam"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			return 0, nil
		},
	}
	w, err := NewWorker(repo, runner, zerolog.Nop(), Config{
		DefaultTimeout:  time.Second,
		DefaultThreads:  2,
		DefaultMemoryMB: 1024,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	payload := `{"reference_path":"/ref.fa","read1_path":"/r1","read2_path":"/r2","output_bam_path":"` + bamPath + `","threads":12,"memory_limit_mb":8192}`
	if err := w.Handle(context.Background(), queue.Message{
		JobID:   created.ID.String(),
		Payload: json.RawMessage(payload),
	}); err != nil {
		t.Fatal(err)
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if len(runner.call) != 2 {
		t.Fatalf("calls=%d", len(runner.call))
	}
	if !strings.HasPrefix(strings.Join(runner.call[0].args, " "), "mem -t 12 -o ") || !strings.Contains(strings.Join(runner.call[0].args, " "), " /ref.fa /r1 /r2") {
		t.Fatalf("bwa args: %v", runner.call[0].args)
	}
	if !strings.HasPrefix(strings.Join(runner.call[1].args, " "), "sort -@ 12 -m 8192M -o "+bamPath+" ") {
		t.Fatalf("samtools args: %v", runner.call[1].args)
	}
}

func TestWorkerHandle_DeterministicChecksum(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	makeJob := func() uuid.UUID {
		j, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
		if err != nil {
			t.Fatal(err)
		}
		return j.ID
	}
	bamPath := filepath.Join(t.TempDir(), "stable.bam")
	runner := &fakeRunner{
		run: func(_ context.Context, name string, _ ...string) (int, error) {
			if name == "samtools" {
				if err := os.WriteFile(bamPath, []byte("same-input-same-output"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			return 0, nil
		},
	}
	w, err := NewWorker(repo, runner, zerolog.Nop(), Config{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	run := func(jobID uuid.UUID) string {
		payload := `{"reference_path":"/ref.fa","read1_path":"/r1","read2_path":"/r2","output_bam_path":"` + bamPath + `"}`
		if err := w.Handle(context.Background(), queue.Message{
			JobID:   jobID.String(),
			Payload: json.RawMessage(payload),
		}); err != nil {
			t.Fatal(err)
		}
		updated, err := repo.GetByID(context.Background(), jobID)
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		if err := json.Unmarshal(updated.OutputRef, &out); err != nil {
			t.Fatal(err)
		}
		sum, _ := out["checksum_sha256"].(string)
		return sum
	}

	sum1 := run(makeJob())
	sum2 := run(makeJob())
	if sum1 == "" || sum2 == "" || sum1 != sum2 {
		t.Fatalf("checksum mismatch sum1=%q sum2=%q", sum1, sum2)
	}
}

func TestWorkerHandle_TimeoutFailure(t *testing.T) {
	t.Parallel()
	repo := newRecordingRepo()
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	w, err := NewWorker(repo, &fakeRunner{
		run: func(ctx context.Context, _ string, _ ...string) (int, error) {
			<-ctx.Done()
			return -1, ctx.Err()
		},
	}, zerolog.Nop(), Config{DefaultTimeout: 20 * time.Millisecond}, nil)
	if err != nil {
		t.Fatal(err)
	}

	payload := `{"reference_path":"/ref.fa","read1_path":"/r1","read2_path":"/r2","output_bam_path":"/tmp/out.bam"}`
	err = w.Handle(context.Background(), queue.Message{
		JobID:   created.ID.String(),
		Payload: json.RawMessage(payload),
	})
	if err != nil {
		t.Fatalf("expected nil handle error after terminal persistence, got %v", err)
	}
	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != job.StatusFailed || updated.Stage != StageAlignmentFailed {
		t.Fatalf("job %+v", updated)
	}
	if !bytes.Contains(updated.OutputRef, []byte(`"error":"context deadline exceeded"`)) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
}

func TestWorkerHandle_FailureCases_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		payload         string
		runner          func(ctx context.Context, name string, args ...string) (int, error)
		expectHandleErr bool
		expectStage     string
		expectStatus    job.Status
		expectOutputErr bool
	}{
		{
			name:            "invalid payload missing required fields",
			payload:         `{"reference_path":"/ref.fa","read1_path":"/r1"}`,
			runner:          func(_ context.Context, _ string, _ ...string) (int, error) { return 0, nil },
			expectHandleErr: true,
			expectStage:     StageAlignmentFailed,
			expectStatus:    job.StatusFailed,
			expectOutputErr: true,
		},
		{
			name:    "runner non-zero exit persists failure",
			payload: `{"reference_path":"/ref.fa","read1_path":"/r1","read2_path":"/r2","output_bam_path":"` + `/tmp/out.bam` + `"}`,
			runner: func(_ context.Context, name string, _ ...string) (int, error) {
				if name == "bwa" {
					return 2, errors.New("bwa failed")
				}
				return 0, nil
			},
			expectHandleErr: false,
			expectStage:     StageAlignmentFailed,
			expectStatus:    job.StatusFailed,
			expectOutputErr: true,
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
			w, err := NewWorker(repo, &fakeRunner{run: tt.runner}, zerolog.Nop(), Config{DefaultTimeout: time.Second}, nil)
			if err != nil {
				t.Fatal(err)
			}

			handleErr := w.Handle(context.Background(), queue.Message{
				JobID:   created.ID.String(),
				Payload: json.RawMessage(tt.payload),
			})
			if tt.expectHandleErr && handleErr == nil {
				t.Fatalf("expected handle error")
			}
			if !tt.expectHandleErr && handleErr != nil {
				t.Fatalf("unexpected handle error: %v", handleErr)
			}

			updated, err := repo.GetByID(context.Background(), created.ID)
			if err != nil {
				t.Fatal(err)
			}
			if updated.Stage != tt.expectStage || updated.Status != tt.expectStatus {
				t.Fatalf("job %+v", updated)
			}
			if tt.expectOutputErr && !bytes.Contains(updated.OutputRef, []byte(`"error"`)) {
				t.Fatalf("output_ref %s", updated.OutputRef)
			}
		})
	}
}
