package fastqc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/stub"
	"github.com/AminN77/senju/backend/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeRunner struct {
	run func(ctx context.Context, args ...string) (int, error)
}

func (f fakeRunner) Run(ctx context.Context, args ...string) (int, error) {
	return f.run(ctx, args...)
}

func TestWorkerHandle_SuccessStoresArtifactsAndLogs(t *testing.T) {
	t.Parallel()
	repo := stub.New()
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
	}, logger, Config{DefaultTimeout: time.Second})
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
	repo := stub.New()
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
	}, zerolog.Nop(), Config{DefaultTimeout: 20 * time.Millisecond})
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
	repo := stub.New()
	worker, err := NewWorker(repo, fakeRunner{
		run: func(_ context.Context, _ ...string) (int, error) { return 0, nil },
	}, zerolog.Nop(), Config{})
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
