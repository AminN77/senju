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
- Security baseline (JWT/RBAC, secret rotation, CI scanning): `docs/security.md`
- Reliability controls (checkpoint/restart, recovery assumptions): `docs/reliability.md`
- Performance qualification and regression gates: `docs/performance.md`
- PR template: `.github/pull_request_template.md`
- Issue templates: `.github/ISSUE_TEMPLATE/`
- Code ownership: `CODEOWNERS`

## Frontend (UI Foundation)

- App location: `frontend/` (Next.js App Router + React + TypeScript strict + pnpm)
- Frontend quick start:
  - `docker compose up -d --build frontend api`
  - open `http://localhost:3001` (or `FRONTEND_PORT`)
  - `docker compose run --rm frontend pnpm lint`
  - `docker compose run --rm frontend pnpm typecheck`
  - `docker compose run --rm frontend pnpm build`
- Frontend engineering docs: `docs/frontend/README.md`
- Design principles: `docs/frontend/design-principles.md`
- Design system and tokens: `docs/frontend/design-system.md`
- Color palette: `docs/frontend/color-palette.md`
- Information architecture: `docs/frontend/information-architecture.md`
- Accessibility (WCAG 2.1 AA baseline): `docs/frontend/accessibility.md`
- Component standards: `docs/frontend/component-standards.md`
- IP guardrails for UX references: `docs/frontend/ip-and-references.md`
- Phase roadmap: `docs/frontend/roadmap.md`

## Architecture decisions

- ADR index: `docs/adr/README.md`
- Current accepted decisions:
  - `docs/adr/0001-queue-system-strategy.md`
  - `docs/adr/0002-database-split-postgres-clickhouse.md`
  - `docs/adr/0003-workflow-orchestration-strategy.md`
  - `docs/adr/0005-frontend-framework.md`
  - `docs/adr/0006-frontend-design-system.md`
  - `docs/adr/0007-frontend-repo-layout.md`

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
- ML impact baseline API (train/predict/persistence): `docs/api/ml-impact.md`
