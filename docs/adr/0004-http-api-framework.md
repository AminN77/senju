# ADR-0004: HTTP API framework (Gin) and control-plane contracts

- **Status:** Accepted
- **Date:** 2026-04-10
- **Owners:** Platform team
- **Decision scope:** Public HTTP API for the genomic data platform (control plane), consistent with the proposal stack and ADR-0001–0003.

## Context

The [Genomic Data Platform Proposal](../../proposals/Genomic%20Data%20Platform%20Proposal.pdf) specifies:

- **Backend:** Go
- **Framework:** Fiber **or** Gin
- **API:** REST (optional GraphQL later)
- **Observability:** metrics and logs (Prometheus/Grafana/Loki in later phases)

ADR-0001–0003 define **NATS**, **PostgreSQL + ClickHouse**, **MinIO**, and **custom Go orchestration** with job state in Postgres. The HTTP API must remain a thin control plane that does not contradict those boundaries.

## Decision

1. **Framework:** Use **[Gin](https://github.com/gin-gonic/gin)** for the Go HTTP API (chosen over Fiber for ecosystem familiarity and middleware patterns; either matches the proposal).
2. **Style:** **REST** with **OpenAPI 3** as the contract source of truth for documented routes.
3. **Health:** Expose **`GET /health/live`** and **`GET /health/ready`** for orchestration; readiness validates dependencies that mirror the proposal and ADRs:
   - PostgreSQL (metadata / jobs per ADR-0002)
   - ClickHouse (analytics per ADR-0002)
   - MinIO (object storage in the proposal)
   - NATS (messaging per ADR-0001)
4. **Version:** Expose **`GET /version`** with build metadata (supports reproducibility in the proposal).
5. **GraphQL:** Out of scope until an explicit ADR and milestone; REST remains canonical.

## Alternatives considered

1. **Fiber** — Pros: similar performance. Cons: duplicate ecosystem vs Gin; either satisfies the proposal.
2. **`net/http` only** — Pros: zero framework deps. Cons: diverges from the stated “Fiber / Gin” stack; more boilerplate for middleware, auth, and versioning later.
3. **GraphQL first** — Rejected for MVP; proposal lists REST first.

## Trade-offs

- **Gain:** Alignment with the proposal, clear path for middleware (auth, request IDs, metrics).
- **Cost:** One framework dependency and learning curve for contributors.

## Risks

- Framework churn if Gin were abandoned — mitigated by keeping handlers thin and domain logic outside HTTP layer.

## Migration path

- If we standardize on Fiber later, introduce an adapter layer and migrate routes incrementally (same OpenAPI paths and status codes).

## Consequences

The control-plane API is documented (OpenAPI), observable at the process level (health), and implementable with the same stack as the proposal and ADR-0001–0003.
