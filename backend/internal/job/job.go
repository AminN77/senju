// Package job defines job metadata types and persistence boundaries for the control plane.
package job

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a job does not exist.
var ErrNotFound = errors.New("job: not found")

// Status is the lifecycle state of a job (matches DB CHECK constraint).
type Status string

const (
	// StatusPending means the job is queued and not yet started.
	StatusPending Status = "pending"
	// StatusRunning means the job is actively executing a stage.
	StatusRunning Status = "running"
	// StatusSucceeded means the job finished without a terminal failure.
	StatusSucceeded Status = "succeeded"
	// StatusFailed means the job ended in error.
	StatusFailed Status = "failed"
	// StatusCancelled means the job was stopped before completion.
	StatusCancelled Status = "cancelled"
)

// Job is transactional metadata for a genomic pipeline job (see ADR-0002).
type Job struct {
	ID          uuid.UUID
	Status      Status
	Stage       string
	InputRef    json.RawMessage
	OutputRef   json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// CreateParams holds fields required to insert a new job row.
type CreateParams struct {
	Status      Status
	Stage       string
	InputRef    json.RawMessage
	OutputRef   json.RawMessage
	StartedAt   *time.Time
	CompletedAt *time.Time
}
