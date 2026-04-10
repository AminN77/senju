# senju
 Genomic Data Processing &amp; Variant Analysis Platform

## Quick start

- Copy env template: `cp .env.example .env`
- Start local stack: `docker compose up -d --build`
- View setup and verification: `docs/setup.md`

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
