// Package fastqc implements the FastQC pipeline stage worker.
package fastqc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	// StageFastQCRunning marks active FastQC stage execution.
	StageFastQCRunning = "fastqc_running"
	// StageFastQCSucceeded marks successful FastQC stage completion.
	StageFastQCSucceeded = "fastqc_succeeded"
	// StageFastQCFailed marks terminal FastQC stage failure.
	StageFastQCFailed = "fastqc_failed"
)

// CommandRunner executes the FastQC command and returns process exit code.
type CommandRunner interface {
	Run(ctx context.Context, args ...string) (int, error)
}

// ExecRunner executes commands with os/exec.
type ExecRunner struct {
}

// Run executes one process and returns either zero or the process exit code.
func (r ExecRunner) Run(ctx context.Context, args ...string) (int, error) {
	// #nosec G204 -- command binary is fixed to "fastqc"; args are validated stage inputs.
	cmd := exec.CommandContext(ctx, "fastqc", args...)
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), err
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return -1, err
	}
	return -1, err
}

// Config controls FastQC worker behavior.
type Config struct {
	DefaultTimeout time.Duration
}

// Worker executes FastQC for queue messages and persists job stage state.
type Worker struct {
	repo   job.Repository
	runner CommandRunner
	log    zerolog.Logger
	cfg    Config
}

// NewWorker creates a FastQC stage worker.
func NewWorker(repo job.Repository, runner CommandRunner, log zerolog.Logger, cfg Config) (*Worker, error) {
	if repo == nil {
		return nil, errors.New("fastqc worker: repo is nil")
	}
	if runner == nil {
		return nil, errors.New("fastqc worker: runner is nil")
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 15 * time.Minute
	}
	return &Worker{repo: repo, runner: runner, log: log, cfg: cfg}, nil
}

type payload struct {
	InputPath      string `json:"input_path"`
	OutputDir      string `json:"output_dir"`
	ReportHTMLURI  string `json:"report_html_uri,omitempty"`
	ReportZIPURI   string `json:"report_zip_uri,omitempty"`
	TimeoutSeconds int64  `json:"timeout_seconds,omitempty"`
}

// Handle is compatible with queue handlers and executes one FastQC job.
func (w *Worker) Handle(ctx context.Context, msg queue.Message) error {
	jobID, err := uuid.Parse(msg.JobID)
	if err != nil {
		return fmt.Errorf("fastqc worker: invalid job id: %w", err)
	}

	p, err := decodePayload(msg.Payload)
	if err != nil {
		return err
	}

	started := time.Now().UTC()
	if _, err := w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:    job.StatusRunning,
		Stage:     StageFastQCRunning,
		StartedAt: &started,
	}); err != nil {
		return fmt.Errorf("fastqc worker: mark running: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, resolveTimeout(w.cfg.DefaultTimeout, p.TimeoutSeconds))
	defer cancel()

	stageStart := time.Now()
	exitCode, runErr := w.runner.Run(runCtx, "--outdir", p.OutputDir, p.InputPath)
	duration := time.Since(stageStart)
	completed := time.Now().UTC()

	status := job.StatusSucceeded
	stage := StageFastQCSucceeded
	if runErr != nil {
		status = job.StatusFailed
		stage = StageFastQCFailed
	}

	outRef, marshalErr := buildOutputRef(exitCode, duration, p, runErr)
	if marshalErr != nil {
		return fmt.Errorf("fastqc worker: output_ref: %w", marshalErr)
	}

	if _, err := w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:      status,
		Stage:       stage,
		CompletedAt: &completed,
		OutputRef:   outRef,
	}); err != nil {
		return fmt.Errorf("fastqc worker: mark completion: %w", err)
	}

	w.log.Info().
		Str("job_id", msg.JobID).
		Str("stage", "fastqc").
		Int("exit_code", exitCode).
		Dur("duration", duration).
		Msg("pipeline stage completed")

	return runErr
}

func decodePayload(raw json.RawMessage) (payload, error) {
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("fastqc worker: decode payload: %w", err)
	}
	if p.InputPath == "" {
		return p, errors.New("fastqc worker: payload input_path is required")
	}
	if p.OutputDir == "" {
		return p, errors.New("fastqc worker: payload output_dir is required")
	}
	return p, nil
}

func resolveTimeout(defaultTimeout time.Duration, timeoutSeconds int64) time.Duration {
	if timeoutSeconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func buildOutputRef(exitCode int, duration time.Duration, p payload, runErr error) (json.RawMessage, error) {
	m := map[string]any{
		"kind":        "fastqc_report_v1",
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
		"artifacts": map[string]any{
			"html_uri": p.ReportHTMLURI,
			"zip_uri":  p.ReportZIPURI,
		},
	}
	if runErr != nil {
		m["error"] = runErr.Error()
	}
	return json.Marshal(m)
}
