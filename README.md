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

## Pipeline queue

- NATS/JetStream queue package with retry + dead-letter semantics: `docs/pipeline/nats-queue.md`
- FastQC worker stage implementation: `backend/internal/pipeline/fastqc`
- BWA + SAMtools alignment worker stage implementation: `backend/internal/pipeline/alignment`
- GATK variant-calling worker stage implementation: `backend/internal/pipeline/gatk`
- Observability baseline (Prometheus, Loki, Grafana dashboard): `docs/observability.md`

## Variant analytics storage

- ClickHouse variant schema and VCF ingestion loader: `docs/data/variants.md`
- Variant query API with filters and pagination: `docs/api/variant-query.md`
- Pipeline orchestration API (create/run/status/outputs): `docs/api/orchestration.md`
