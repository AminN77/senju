package postgres

import (
	"context"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/db/jobdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements [job.Repository] using PostgreSQL and sqlc-generated queries.
type Repository struct {
	q *jobdb.Queries
}

// NewRepository returns a Postgres-backed job repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: jobdb.New(pool)}
}

// Create implements [job.Repository].
func (r *Repository) Create(ctx context.Context, p job.CreateParams) (*job.Job, error) {
	row, err := r.q.CreateJob(ctx, mapCreateParams(p))
	if err != nil {
		return nil, err
	}
	return rowToJob(row)
}

// GetByID implements [job.Repository].
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	row, err := r.q.GetJobByID(ctx, uuidToPgUUID(id))
	if err != nil {
		return nil, mapNotFound(err)
	}
	return rowToJob(row)
}

// Update implements [job.Repository].
func (r *Repository) Update(ctx context.Context, id uuid.UUID, p job.UpdateParams) (*job.Job, error) {
	row, err := r.q.UpdateJob(ctx, mapUpdateParams(id, p))
	if err != nil {
		return nil, mapNotFound(err)
	}
	return rowToJob(row)
}
