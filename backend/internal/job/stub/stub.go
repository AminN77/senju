// Package stub provides an in-memory [job.Repository] for unit tests.
package stub

import (
	"context"
	"encoding/json"
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

func cloneRawMessage(in json.RawMessage) json.RawMessage {
	if len(in) == 0 {
		return nil
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := *t
	return &v
}

func cloneJob(j job.Job) job.Job {
	return job.Job{
		ID:          j.ID,
		Status:      j.Status,
		Stage:       j.Stage,
		InputRef:    cloneRawMessage(j.InputRef),
		OutputRef:   cloneRawMessage(j.OutputRef),
		CreatedAt:   j.CreatedAt,
		UpdatedAt:   j.UpdatedAt,
		StartedAt:   cloneTimePtr(j.StartedAt),
		CompletedAt: cloneTimePtr(j.CompletedAt),
	}
}

// Create implements [job.Repository].
func (r *Repository) Create(ctx context.Context, p job.CreateParams) (*job.Job, error) {
	_ = ctx
	now := time.Now().UTC()
	j := job.Job{
		ID:          uuid.New(),
		Status:      p.Status,
		Stage:       p.Stage,
		InputRef:    cloneRawMessage(p.InputRef),
		OutputRef:   cloneRawMessage(p.OutputRef),
		CreatedAt:   now,
		UpdatedAt:   now,
		StartedAt:   cloneTimePtr(p.StartedAt),
		CompletedAt: cloneTimePtr(p.CompletedAt),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := cloneJob(j)
	r.jobs[j.ID] = stored
	out := cloneJob(stored)
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
	out := cloneJob(j)
	return &out, nil
}

// Update implements [job.Repository].
func (r *Repository) Update(ctx context.Context, id uuid.UUID, p job.UpdateParams) (*job.Job, error) {
	_ = ctx
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return nil, job.ErrNotFound
	}
	j.Status = p.Status
	j.Stage = p.Stage
	if len(p.OutputRef) > 0 {
		j.OutputRef = cloneRawMessage(p.OutputRef)
	}
	if p.StartedAt != nil {
		j.StartedAt = cloneTimePtr(p.StartedAt)
	}
	if p.CompletedAt != nil {
		j.CompletedAt = cloneTimePtr(p.CompletedAt)
	}
	j.UpdatedAt = time.Now().UTC()
	r.jobs[id] = cloneJob(j)
	out := cloneJob(j)
	return &out, nil
}
