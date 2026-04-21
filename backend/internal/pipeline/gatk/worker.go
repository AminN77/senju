// Package gatk implements the GATK variant-calling pipeline worker.
package gatk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/queue"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

const (
	// StageGATKRunning marks active GATK execution.
	StageGATKRunning = "gatk_running"
	// StageGATKSucceeded marks successful GATK completion.
	StageGATKSucceeded = "gatk_succeeded"
	// StageGATKFailed marks terminal GATK stage failure.
	StageGATKFailed = "gatk_failed"

	// ErrorClassTool indicates tool/runtime command failure.
	ErrorClassTool = "tool"
	// ErrorClassInfrastructure indicates timeout/infra/runtime environment failure.
	ErrorClassInfrastructure = "infrastructure"
	// ErrorClassConfiguration indicates invalid payload or bad stage configuration.
	ErrorClassConfiguration = "configuration"
)

// CommandRunner executes GATK command invocations and returns process exit code.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (int, error)
}

// ExecRunner executes command binaries through os/exec.
type ExecRunner struct{}

// Run executes one command and returns either zero or process exit code.
func (ExecRunner) Run(ctx context.Context, name string, args ...string) (int, error) {
	// #nosec G204 -- command binary/args are controlled by validated stage payload/config.
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

// Metrics exports GATK stage metrics to Prometheus.
type Metrics struct {
	duration *prometheus.HistogramVec
	total    *prometheus.CounterVec
}

// NewMetrics returns Prometheus collectors for GATK stage observability.
func NewMetrics() *Metrics {
	return &Metrics{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "senju_pipeline_stage_duration_seconds",
			Help:    "Duration of pipeline stage executions by stage and outcome.",
			Buckets: prometheus.DefBuckets,
		}, []string{"stage", "outcome"}),
		total: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "senju_pipeline_stage_total",
			Help: "Total pipeline stage executions by stage, outcome, and classified error type.",
		}, []string{"stage", "outcome", "error_class"}),
	}
}

// Collectors returns the metrics collectors for registry registration.
func (m *Metrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{m.duration, m.total}
}

// Observe records duration and terminal status classification.
func (m *Metrics) Observe(stage, outcome, errorClass string, duration time.Duration) {
	m.duration.WithLabelValues(stage, outcome).Observe(duration.Seconds())
	m.total.WithLabelValues(stage, outcome, errorClass).Inc()
}

// Config controls GATK worker defaults.
type Config struct {
	DefaultTimeout  time.Duration
	DefaultThreads  int
	DefaultMemoryMB int
}

// Worker executes GATK stage for queue jobs and persists stage transitions.
type Worker struct {
	repo    job.Repository
	runner  CommandRunner
	log     zerolog.Logger
	cfg     Config
	metrics *Metrics
}

// NewWorker creates a GATK worker with defaults and optional metrics.
func NewWorker(repo job.Repository, runner CommandRunner, log zerolog.Logger, cfg Config, metrics *Metrics) (*Worker, error) {
	if repo == nil {
		return nil, errors.New("gatk worker: repo is nil")
	}
	if runner == nil {
		return nil, errors.New("gatk worker: runner is nil")
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 90 * time.Minute
	}
	if cfg.DefaultThreads <= 0 {
		cfg.DefaultThreads = 4
	}
	if cfg.DefaultMemoryMB <= 0 {
		cfg.DefaultMemoryMB = 4096
	}
	return &Worker{repo: repo, runner: runner, log: log, cfg: cfg, metrics: metrics}, nil
}

type payload struct {
	ReferencePath  string `json:"reference_path"`
	InputBAMPath   string `json:"input_bam_path"`
	OutputVCFPath  string `json:"output_vcf_path"`
	OutputVCFURI   string `json:"output_vcf_uri,omitempty"`
	TimeoutSeconds int64  `json:"timeout_seconds,omitempty"`
	Threads        int    `json:"threads,omitempty"`
	MemoryLimitMB  int    `json:"memory_limit_mb,omitempty"`
}

// Handle executes one GATK stage invocation for the queue message job.
func (w *Worker) Handle(ctx context.Context, msg queue.Message) error {
	jobID, err := uuid.Parse(msg.JobID)
	if err != nil {
		return fmt.Errorf("gatk worker: invalid job id: %w", err)
	}
	p, err := decodePayload(msg.Payload)
	if err != nil {
		if perr := w.persistFailure(ctx, jobID, -1, payload{}, 0, 0, 0, ErrorClassConfiguration, err); perr != nil {
			return fmt.Errorf("gatk worker: persist configuration failure: %w", perr)
		}
		w.observe("failure", ErrorClassConfiguration, 0)
		return nil
	}

	started := time.Now().UTC()
	if _, err := w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:    job.StatusRunning,
		Stage:     StageGATKRunning,
		StartedAt: &started,
	}); err != nil {
		return fmt.Errorf("gatk worker: mark running: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, resolveTimeout(w.cfg.DefaultTimeout, p.TimeoutSeconds))
	defer cancel()

	stageStart := time.Now()
	threads := choosePositive(p.Threads, w.cfg.DefaultThreads)
	memMB := choosePositive(p.MemoryLimitMB, w.cfg.DefaultMemoryMB)
	exitCode, runErr := w.runPipeline(runCtx, p, threads, memMB)
	duration := time.Since(stageStart)

	if runErr != nil {
		class := classifyError(runErr, exitCode)
		if err := w.persistFailure(ctx, jobID, exitCode, p, threads, memMB, duration, class, runErr); err != nil {
			return fmt.Errorf("gatk worker: persist failure: %w", err)
		}
		w.observe("failure", class, duration)
		w.log.Info().Str("job_id", msg.JobID).Str("stage", "gatk").Int("exit_code", exitCode).Dur("duration", duration).Msg("pipeline stage completed")
		return nil
	}

	outRef, err := buildOutputRef(exitCode, duration, p, threads, memMB, "", "")
	if err != nil {
		return fmt.Errorf("gatk worker: output_ref: %w", err)
	}
	completed := time.Now().UTC()
	if _, err := w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:      job.StatusSucceeded,
		Stage:       StageGATKSucceeded,
		CompletedAt: &completed,
		OutputRef:   outRef,
	}); err != nil {
		return fmt.Errorf("gatk worker: mark success: %w", err)
	}
	w.observe("success", "", duration)
	w.log.Info().Str("job_id", msg.JobID).Str("stage", "gatk").Int("exit_code", exitCode).Dur("duration", duration).Msg("pipeline stage completed")
	return nil
}

func (w *Worker) runPipeline(ctx context.Context, p payload, threads, memMB int) (int, error) {
	args := []string{
		"--java-options", fmt.Sprintf("-Xmx%dM", memMB),
		"HaplotypeCaller",
		"-R", p.ReferencePath,
		"-I", p.InputBAMPath,
		"-O", p.OutputVCFPath,
		"--native-pair-hmm-threads", fmt.Sprintf("%d", threads),
	}
	return w.runner.Run(ctx, "gatk", args...)
}

func (w *Worker) persistFailure(ctx context.Context, jobID uuid.UUID, exitCode int, p payload, threads, memMB int, duration time.Duration, errorClass string, runErr error) error {
	outRef, err := buildOutputRef(exitCode, duration, p, threads, memMB, errorClass, runErr.Error())
	if err != nil {
		return err
	}
	completed := time.Now().UTC()
	_, err = w.repo.Update(ctx, jobID, job.UpdateParams{
		Status:      job.StatusFailed,
		Stage:       StageGATKFailed,
		CompletedAt: &completed,
		OutputRef:   outRef,
	})
	return err
}

func (w *Worker) observe(outcome, class string, duration time.Duration) {
	if w.metrics == nil {
		return
	}
	w.metrics.Observe("gatk", outcome, class, duration)
}

func decodePayload(raw json.RawMessage) (payload, error) {
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("gatk worker: decode payload: %w", err)
	}
	if p.ReferencePath == "" {
		return p, errors.New("gatk worker: payload reference_path is required")
	}
	if p.InputBAMPath == "" {
		return p, errors.New("gatk worker: payload input_bam_path is required")
	}
	if p.OutputVCFPath == "" {
		return p, errors.New("gatk worker: payload output_vcf_path is required")
	}
	return p, nil
}

func choosePositive(v, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func resolveTimeout(defaultTimeout time.Duration, timeoutSeconds int64) time.Duration {
	if timeoutSeconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func classifyError(err error, exitCode int) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "payload"), strings.Contains(msg, "decode"), strings.Contains(msg, "required"), strings.Contains(msg, "invalid job id"):
		return ErrorClassConfiguration
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return ErrorClassInfrastructure
	case exitCode > 0:
		return ErrorClassTool
	default:
		return ErrorClassInfrastructure
	}
}

func buildOutputRef(exitCode int, duration time.Duration, p payload, threads, memMB int, errorClass, errorMsg string) (json.RawMessage, error) {
	m := map[string]any{
		"kind":        "gatk_vcf_v1",
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
		"limits": map[string]any{
			"threads":         threads,
			"memory_limit_mb": memMB,
		},
		"artifacts": map[string]any{
			"vcf_path": p.OutputVCFPath,
			"vcf_uri":  p.OutputVCFURI,
		},
	}
	if errorClass != "" {
		m["error_class"] = errorClass
	}
	if errorMsg != "" {
		m["error"] = errorMsg
	}
	return json.Marshal(m)
}
