# Job metadata (PostgreSQL)

Transactional **job** rows live in PostgreSQL per [ADR-0002](../adr/0002-database-split-postgres-clickhouse.md). Pipeline **stage** and lifecycle **status** are tracked here for orchestration ([ADR-0003](../adr/0003-workflow-orchestration-strategy.md)).

## Schema

The `public.jobs` table is defined in [backend/migrations/000001_jobs.up.sql](../../backend/migrations/000001_jobs.up.sql). DDL is schema-qualified (`public.*`) so behavior does not depend on `search_path`. Row `id` values are generated in the application on insert (no `pgcrypto` / extension ownership in this migration). The revision defines `public.jobs_set_updated_at()` for the `updated_at` trigger and removes it in the paired down migration.

- `id` (UUID, primary key)
- `status` (`pending` | `running` | `succeeded` | `failed` | `cancelled`)
- `stage` (text): current pipeline stage label
- `input_ref`, `output_ref` (JSONB, nullable): structured references (e.g. object store URIs)
- `created_at`, `updated_at` (timestamptz): `updated_at` maintained by a trigger
- `started_at`, `completed_at` (timestamptz, nullable)

Indexes support listing recent jobs and filtering active work.

## Migrations

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate) with SQL files under `backend/migrations/`. Embedded copies are available for tooling via `backend/migrations/embed.go`.

Apply from the **backend** module (same DSN rules as the API — `POSTGRES_DSN` or `POSTGRES_HOST` / `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB`):

```bash
cd backend
go run ./cmd/migrate up
go run ./cmd/migrate version
# rollback (development)
go run ./cmd/migrate down
```

## Application code

- Domain types and [`job.Repository`](../../backend/internal/job/repository.go) live in `internal/job/`.
- The PostgreSQL implementation uses **sqlc** + **pgx** (`internal/job/db/jobdb/`, queries in `internal/job/db/queries/`).
- Regenerate sqlc output after changing queries:

  ```bash
  cd backend
  sqlc generate
  ```

## Tracking issue

Implemented for [Issue #5](https://github.com/AminN77/senju/issues/5).
