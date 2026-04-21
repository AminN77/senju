# Local development setup (Issue 3)

This setup bootstraps the core local development services:

- API container (minimal bootstrap service)
- PostgreSQL
- ClickHouse
- MinIO
- NATS

## Prerequisites

- Docker Engine or Docker Desktop
- Docker Compose v2

## Quick start

1. Create local env file:

   ```bash
   cp .env.example .env
   ```

2. Start services:

   ```bash
   docker compose up -d --build
   ```

3. Check service status:

   ```bash
   docker compose ps
   ```

4. Check logs (optional):

   ```bash
   docker compose logs -f --tail=100
   ```

## Database migrations

After Postgres is up, apply SQL migrations so the `jobs` metadata table exists (see [docs/data/jobs.md](./data/jobs.md)):

```bash
cd backend
go run ./cmd/migrate up
```

Use the same environment variables as in `.env` (`POSTGRES_*` or `POSTGRES_DSN`) so the migrate command targets your Compose Postgres instance.

## FASTQ upload metadata API

After migrations and with Postgres reachable, the API can register FASTQ ingestion jobs. See [docs/api/fastq-upload.md](./api/fastq-upload.md).

## Variant query API

Variant retrieval with filters/pagination is documented in [docs/api/variant-query.md](./api/variant-query.md).

## Pipeline orchestration API

Create/run/status/output orchestration endpoints are documented in [docs/api/orchestration.md](./api/orchestration.md).

## Queue semantics (NATS + retries)

Queue retry/dead-letter behavior and env configuration are documented in [docs/pipeline/nats-queue.md](./pipeline/nats-queue.md).

## Observability baseline

Monitoring/logging stack setup (Prometheus, Loki, Grafana) is documented in [docs/observability.md](./observability.md).

## Verification commands

- API:
  - `curl http://localhost:8080/health/live`
  - `curl http://localhost:8080/health/ready` (expects 200 when dependencies are reachable)
  - `curl http://localhost:8080/version`
  - OpenAPI document: `backend/openapi/openapi.yaml`
- Postgres:
  - `docker compose exec postgres pg_isready -U ${POSTGRES_USER:-senju} -d ${POSTGRES_DB:-senju}`
- ClickHouse:
  - `curl http://localhost:8123/ping`
- MinIO:
  - `curl http://localhost:9001/minio/health/live`
  - Console: `http://localhost:9002`
- NATS monitoring:
  - `curl http://localhost:8222/varz`
- Prometheus:
  - `curl http://localhost:9090/api/v1/targets`
- Grafana:
  - `http://localhost:3000` (default from `.env`: `admin` / `admin`)

## Stop and cleanup

- Stop:
  - `docker compose down`
- Stop and remove volumes:
  - `docker compose down -v`
