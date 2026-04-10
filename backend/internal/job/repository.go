package job

import (
	"context"

	"github.com/google/uuid"
)

// Repository persists and loads job metadata. Implementations may target PostgreSQL or other stores.
type Repository interface {
	Create(ctx context.Context, p CreateParams) (*Job, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	Update(ctx context.Context, id uuid.UUID, p UpdateParams) (*Job, error)
}
