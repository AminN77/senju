# senju
 Genomic Data Processing &amp; Variant Analysis Platform

## Quick start

- Copy env template: `cp .env.example .env`
- Start local stack: `docker compose up -d --build`
- View setup and verification: `docs/setup.md`

## HTTP API (skeleton)

- **Backend standards:** `docs/backend-engineering.md` (Uber style guide, linting, tests, concurrency)
- **Stack:** Go + **Gin**, REST, OpenAPI (`docs/adr/0004-http-api-framework.md`)
- `GET /health/live` — liveness
- `GET /health/ready` — readiness (Postgres, ClickHouse, MinIO, NATS — aligned with ADRs and the platform proposal)
- `GET /version` — JSON build metadata
- OpenAPI: `backend/openapi/openapi.yaml`

## Engineering governance

- Contribution standards: `docs/contributing.md`
- PR template: `.github/pull_request_template.md`
- Issue templates: `.github/ISSUE_TEMPLATE/`
- Code ownership: `CODEOWNERS`

## Architecture decisions

- ADR index: `docs/adr/README.md`
- Current accepted decisions:
  - `docs/adr/0001-queue-system-strategy.md`
  - `docs/adr/0002-database-split-postgres-clickhouse.md`
  - `docs/adr/0003-workflow-orchestration-strategy.md`
