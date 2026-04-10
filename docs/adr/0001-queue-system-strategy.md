# ADR-0001: Queue system strategy (NATS first)

- **Status:** Accepted
- **Date:** 2026-04-10
- **Owners:** Platform team
- **Decision scope:** Job dispatch and worker communication for genomics pipeline stages.

## Context

The platform needs asynchronous execution for pipeline stages (QC, alignment, variant calling), retries, and worker scaling. The initial team is small, so operational simplicity is a priority during MVP.

## Decision

Use **NATS** as the initial queue and messaging backbone for job dispatch and worker communication.

## Alternatives considered

1. **RabbitMQ**
   - Pros: mature routing features, strong queue semantics.
   - Cons: higher operational overhead for current team size.
2. **Kafka**
   - Pros: durable event streams at large scale.
   - Cons: heavy operational complexity for MVP workload.
3. **Redis streams**
   - Pros: simple for small systems.
   - Cons: less aligned with long-term eventing and multi-consumer topology needs.

## Trade-offs

- **Gain:** fast implementation, low operational burden, straightforward local development with Docker.
- **Cost:** fewer advanced stream-processing patterns than Kafka.
- **Constraint:** needs deliberate retry/dead-letter design in application layer.

## Risks

- Message handling bugs may cause duplicate work or unacked jobs.
- Insufficient observability could hide queue backlogs until late.

## Mitigations

- Enforce explicit ack/retry/dead-letter behavior per job type.
- Add queue depth, retry count, and processing latency metrics.
- Build idempotency in worker stages (safe re-processing).

## Migration path

1. Keep queue interactions behind a `queue` interface in Go.
2. Separate message contracts from transport implementation.
3. Add adapter for future Kafka/RabbitMQ backend when scale or compliance requires.
4. Run dual-publish shadow mode before full transport cutover.

## Consequences

NATS is the default queue for MVP and early production-like environments. Future migration remains possible without redesigning business logic if transport boundaries are respected.
