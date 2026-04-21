// Package alignment implements the BWA + SAMtools alignment pipeline worker.
package alignment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/pipeline/stagemetrics"
	"github.com/AminN77/senju/backend/internal/queue"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	// StageAlignmentRunning marks active alignment execution.
	StageAlignmentRunning = "alignment_running"
	// StageAlignmentSucceeded marks successful alignment completion.
	StageAlignmentSucceeded = "alignment_succeeded"
	// StageAlignmentFailed marks terminal alignment failure.
	StageAlignmentFailed = "alignment_failed"
)

// CommandRunner executes external alignment commands and returns process exit code.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (int, error)
}

// ExecRunner executes command binaries via os/exec.
type ExecRunner struct{}

// Run executes a command and returns zero or process exit code.
func (ExecRunner) Run(ctx context.Context, name string, args ...string) (int, error) {
	// #nosec G204 -- binary name and args are constructed from validated stage config/payload.
	cmd := exec.CommandContext(ctx, name, args...)
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

// Config controls alignment execution defaults.
type Config struct {
	DefaultTimeout  time.Duration
	DefaultThreads  int
	DefaultMemoryMB int
}

// Worker runs BWA+SAMtools stage and persists BAM metadata in output_ref.
type Worker struct {
	repo    job.Repository
	runner  CommandRunner
	log     zerolog.Logger
	cfg     Config
	metrics *stagemetrics.Metrics
}

// NewWorker creates an alignment worker with sane defaults.
func NewWorker(repo job.Repository, runner CommandRunner, log zerolog.Logger, cfg Config, metrics *stagemetrics.Metrics) (*Worker, error) {
	if repo == nil {
		return nil, errors.New("alignment worker: repo is nil")
	}
	if runner == nil {
		return nil, errors.New("alignment worker: runner is nil")
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 60 * time.Minute
	}
	if cfg.DefaultThreads <= 0 {
		cfg.DefaultThreads = 4
	}
	if cfg.DefaultMemoryMB <= 0 {
		cfg.DefaultMemoryMB = 2048
	}
	return &Worker{repo: repo, runner: runner, log: log, cfg: cfg, metrics: metrics}, nil
}

type payload struct {
	ReferencePath  string `json:"reference_path"`
	Read1Path      string `json:"read1_path"`
	Read2Path      string `json:"read2_path"`
	OutputBAMPath  string `json:"output_bam_path"`
	OutputBAMURI   string `json:"output_bam_uri,omitempty"`
	TimeoutSeconds int64  `json:"timeout_seconds,omitempty"`
	Threads        int    `json:"threads,omitempty"`
	MemoryLimitMB  int    `json:"memory_limit_mb,omitempty"`
}

// Handle executes alignment for one queued job message.
func (w *Worker) Handle(ctx context.Context, msg queue.Message) error {
	jobID, err := uuid.Parse(msg.JobID)
	if err != nil {
		return fmt.Errorf("alignment worker: invalid job id: %w", err)
	}
	p, err := decodePayload(msg.Payload)
	if err != nil {
		if w.metrics != nil {
			w.metrics.Observe("alignment", "failure", "validation", 0)
		}
		_ = w.persistTerminalFailure(ctx, jobID, -1, payload{}, 0, 0, err)
		return err
	}
	started := time.Now().UTC()
	if _, err := w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:    job.StatusRunning,
		Stage:     StageAlignmentRunning,
		StartedAt: &started,
	}); err != nil {
		return fmt.Errorf("alignment worker: mark running: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, resolveTimeout(w.cfg.DefaultTimeout, p.TimeoutSeconds))
	defer cancel()

	stageStart := time.Now()
	threads := choosePositive(p.Threads, w.cfg.DefaultThreads)
	memMB := choosePositive(p.MemoryLimitMB, w.cfg.DefaultMemoryMB)

	exitCode, runErr := w.runPipeline(runCtx, p, threads, memMB)
	duration := time.Since(stageStart)
	completed := time.Now().UTC()

	checksum := ""
	if runErr == nil {
		sum, err := checksumFileSHA256(p.OutputBAMPath)
		if err != nil {
			runErr = fmt.Errorf("alignment worker: checksum bam: %w", err)
			exitCode = -1
		} else {
			checksum = sum
		}
	}

	status := job.StatusSucceeded
	stage := StageAlignmentSucceeded
	outcome := "success"
	errClass := ""
	if runErr != nil {
		status = job.StatusFailed
		stage = StageAlignmentFailed
		outcome = "failure"
		errClass = "tool"
		if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
			errClass = "infrastructure"
		}
	}
	if w.metrics != nil {
		w.metrics.Observe("alignment", outcome, errClass, duration)
	}

	outRef, marshalErr := buildOutputRef(exitCode, duration, checksum, p, threads, memMB, runErr)
	if marshalErr != nil {
		return fmt.Errorf("alignment worker: output_ref: %w", marshalErr)
	}
	if _, err := w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:      status,
		Stage:       stage,
		CompletedAt: &completed,
		OutputRef:   outRef,
	}); err != nil {
		return fmt.Errorf("alignment worker: mark completion: %w", err)
	}

	w.log.Info().
		Str("job_id", msg.JobID).
		Str("stage", "alignment").
		Int("exit_code", exitCode).
		Dur("duration", duration).
		Msg("pipeline stage completed")

	// Terminal stage state is persisted above; acknowledge message to avoid
	// retrying already-finalized jobs.
	return nil
}

func (w *Worker) runPipeline(ctx context.Context, p payload, threads, memMB int) (int, error) {
	tmpSAMPath := filepath.Join(os.TempDir(), "senju-align-"+uuid.NewString()+".sam")
	defer func() { _ = os.Remove(tmpSAMPath) }()

	bwaArgs := []string{
		"mem",
		"-t", fmt.Sprintf("%d", threads),
		"-o", tmpSAMPath,
		p.ReferencePath,
		p.Read1Path,
		p.Read2Path,
	}
	if code, err := w.runner.Run(ctx, "bwa", bwaArgs...); err != nil {
		return code, err
	}

	samtoolsArgs := []string{
		"sort",
		"-@", fmt.Sprintf("%d", threads),
		"-m", fmt.Sprintf("%dM", memMB),
		"-o", p.OutputBAMPath,
		tmpSAMPath,
	}
	code, err := w.runner.Run(ctx, "samtools", samtoolsArgs...)
	return code, err
}

func (w *Worker) persistTerminalFailure(ctx context.Context, jobID uuid.UUID, exitCode int, p payload, threads, memMB int, failure error) error {
	outRef, err := buildOutputRef(exitCode, 0, "", p, threads, memMB, failure)
	if err != nil {
		return err
	}
	completed := time.Now().UTC()
	_, err = w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:      job.StatusFailed,
		Stage:       StageAlignmentFailed,
		CompletedAt: &completed,
		OutputRef:   outRef,
	})
	return err
}

func decodePayload(raw json.RawMessage) (payload, error) {
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("alignment worker: decode payload: %w", err)
	}
	if p.ReferencePath == "" {
		return p, errors.New("alignment worker: payload reference_path is required")
	}
	if p.Read1Path == "" {
		return p, errors.New("alignment worker: payload read1_path is required")
	}
	if p.Read2Path == "" {
		return p, errors.New("alignment worker: payload read2_path is required")
	}
	if p.OutputBAMPath == "" {
		return p, errors.New("alignment worker: payload output_bam_path is required")
	}
	return p, nil
}

func resolveTimeout(defaultTimeout time.Duration, timeoutSeconds int64) time.Duration {
	if timeoutSeconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func choosePositive(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func checksumFileSHA256(path string) (string, error) {
	// #nosec G304 -- path is stage output path from validated worker payload/config.
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func buildOutputRef(exitCode int, duration time.Duration, checksum string, p payload, threads, memMB int, runErr error) (json.RawMessage, error) {
	m := map[string]any{
		"kind":        "alignment_bam_v1",
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
		"limits": map[string]any{
			"threads":         threads,
			"memory_limit_mb": memMB,
		},
		"artifacts": map[string]any{
			"bam_path": p.OutputBAMPath,
			"bam_uri":  p.OutputBAMURI,
		},
	}
	if checksum != "" {
		m["checksum_sha256"] = checksum
	}
	if runErr != nil {
		m["error"] = runErr.Error()
	}
	return json.Marshal(m)
}
