-- name: CreateJob :one
INSERT INTO jobs (
    status,
    stage,
    input_ref,
    output_ref,
    started_at,
    completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, status, stage, input_ref, output_ref, created_at, updated_at, started_at, completed_at;

-- name: GetJobByID :one
SELECT id, status, stage, input_ref, output_ref, created_at, updated_at, started_at, completed_at
FROM jobs
WHERE id = $1;

-- name: UpdateJob :one
UPDATE jobs
SET status = $2,
    stage = $3,
    started_at = COALESCE($4, started_at),
    completed_at = COALESCE($5, completed_at)
WHERE id = $1
RETURNING id, status, stage, input_ref, output_ref, created_at, updated_at, started_at, completed_at;
