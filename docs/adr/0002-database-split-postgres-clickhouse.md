# ADR-0002: Database split (PostgreSQL + ClickHouse)

- **Status:** Accepted
- **Date:** 2026-04-10
- **Owners:** Platform team
- **Decision scope:** Persistence strategy for metadata and analytical genomic variants.

## Context

The platform stores two distinct data shapes:

- transactional metadata (jobs, statuses, users, pipeline state)
- analytical variant data (high-volume query workloads by gene/chromosome/position)

A single database for both concerns would create pressure between OLTP and analytical access patterns.

## Decision

Adopt a **split-database architecture**:

- **PostgreSQL** for transactional metadata and control-plane state.
- **ClickHouse** for variant analytics and large-scale query workloads.

## Alternatives considered

1. **PostgreSQL only**
   - Pros: one datastore, simpler initial setup.
   - Cons: analytical workloads may degrade transactional reliability at scale.
2. **ClickHouse only**
   - Pros: strong analytical performance.
   - Cons: weak fit for transactional workflow orchestration and strict row-level updates.
3. **Elasticsearch for analytics + Postgres for metadata**
   - Pros: flexible search.
   - Cons: added indexing complexity and less direct fit for numeric analytical scans.

## Trade-offs

- **Gain:** each datastore is used for the workload it handles best.
- **Cost:** added operational complexity and ETL/load pipeline between workflow outputs and analytics store.
- **Constraint:** requires clear ownership boundaries to avoid data-model drift.

## Risks

- Data consistency lag between pipeline completion and query availability.
- Schema evolution mismatch between transactional and analytical layers.

## Mitigations

- Define canonical variant ingestion contract and version it.
- Track ingestion lag and failure metrics.
- Apply migration discipline for both Postgres and ClickHouse schemas.

## Migration path

1. Start with minimal schemas for jobs (Postgres) and variants (ClickHouse).
2. Add ingestion loader with idempotent upsert/dedup semantics.
3. Introduce schema versioning and compatibility checks.
4. If needed, add data lake/object-store layer later without changing control-plane model.

## Consequences

The system gains strong query performance and cleaner domain separation at the cost of integration complexity. This is acceptable for platform goals and expected data growth.
