// Package postgres implements [job.Repository] using PostgreSQL (sqlc + pgx).
package postgres

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/db/jobdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func rowToJob(row jobdb.Job) (*job.Job, error) {
	id, err := pgUUIDToUUID(row.ID)
	if err != nil {
		return nil, err
	}
	j := &job.Job{
		ID:     id,
		Status: job.Status(row.Status),
		Stage:  row.Stage,
	}
	if len(row.InputRef) > 0 {
		j.InputRef = json.RawMessage(append([]byte(nil), row.InputRef...))
	}
	if len(row.OutputRef) > 0 {
		j.OutputRef = json.RawMessage(append([]byte(nil), row.OutputRef...))
	}
	if row.CreatedAt.Valid {
		j.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		j.UpdatedAt = row.UpdatedAt.Time
	}
	j.StartedAt = timestamptzPtr(row.StartedAt)
	j.CompletedAt = timestamptzPtr(row.CompletedAt)
	return j, nil
}

func pgUUIDToUUID(u pgtype.UUID) (uuid.UUID, error) {
	if !u.Valid {
		return uuid.Nil, errors.New("invalid uuid")
	}
	return uuid.UUID(u.Bytes), nil
}

func uuidToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func timestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tt := t.Time
	return &tt
}

func timeToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func mapCreateParams(p job.CreateParams) jobdb.CreateJobParams {
	arg := jobdb.CreateJobParams{
		Status: string(p.Status),
		Stage:  p.Stage,
	}
	if len(p.InputRef) > 0 {
		arg.InputRef = append([]byte(nil), p.InputRef...)
	}
	if len(p.OutputRef) > 0 {
		arg.OutputRef = append([]byte(nil), p.OutputRef...)
	}
	arg.StartedAt = timeToTimestamptz(p.StartedAt)
	arg.CompletedAt = timeToTimestamptz(p.CompletedAt)
	return arg
}

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return job.ErrNotFound
	}
	return err
}
