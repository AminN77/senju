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

## Verification commands

- API:
  - `curl http://localhost:8080/`
- Postgres:
  - `docker compose exec postgres pg_isready -U ${POSTGRES_USER:-senju} -d ${POSTGRES_DB:-senju}`
- ClickHouse:
  - `curl http://localhost:8123/ping`
- MinIO:
  - `curl http://localhost:9001/minio/health/live`
  - Console: `http://localhost:9002`
- NATS monitoring:
  - `curl http://localhost:8222/varz`

## Stop and cleanup

- Stop:
  - `docker compose down`
- Stop and remove volumes:
  - `docker compose down -v`
