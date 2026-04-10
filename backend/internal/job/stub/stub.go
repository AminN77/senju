// Package stub provides an in-memory [job.Repository] for unit tests.
package stub

import (
	"context"
	"sync"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/google/uuid"
)

// Repository is a non-persistent store (single-process tests only).
type Repository struct {
	mu   sync.Mutex
	jobs map[uuid.UUID]job.Job
}

// New returns an empty stub repository.
func New() *Repository {
	return &Repository{jobs: make(map[uuid.UUID]job.Job)}
}

// Create implements [job.Repository].
func (r *Repository) Create(ctx context.Context, p job.CreateParams) (*job.Job, error) {
	_ = ctx
	now := time.Now().UTC()
	j := job.Job{
		ID:          uuid.New(),
		Status:      p.Status,
		Stage:       p.Stage,
		InputRef:    p.InputRef,
		OutputRef:   p.OutputRef,
		CreatedAt:   now,
		UpdatedAt:   now,
		StartedAt:   p.StartedAt,
		CompletedAt: p.CompletedAt,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[j.ID] = j
	out := j
	return &out, nil
}

// GetByID implements [job.Repository].
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return nil, job.ErrNotFound
	}
	out := j
	return &out, nil
}

// UpdateStatusStage implements [job.Repository].
func (r *Repository) UpdateStatusStage(ctx context.Context, id uuid.UUID, status job.Status, stage string) (*job.Job, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return nil, job.ErrNotFound
	}
	j.Status = status
	j.Stage = stage
	r.jobs[id] = j
	out := j
	return &out, nil
}
