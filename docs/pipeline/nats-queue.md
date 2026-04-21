# NATS queue integration

Issue #9 introduces a JetStream-backed queue package for workflow dispatch with retry and dead-letter behavior.

## Package

- Path: `backend/internal/queue`
- Implementation: `NATSQueue` (`nats.go`)
- Contract: `Queue` interface (`queue.go`)

## Behavior

- `Enqueue` publishes one job message to the main work subject.
- `Consume` pulls from a durable JetStream consumer and calls your handler.
- Handler success (`nil`) => message is acknowledged.
- Handler failure (`error`) => message is retried with exponential backoff.
- After retry cap is exceeded => message is published to dead-letter subject and acknowledged (no silent loss).

## Configuration

Queue settings are loaded from environment variables in `backend/internal/config`:

- `QUEUE_STREAM_NAME` (default `jobs_stream`)
- `QUEUE_SUBJECT` (default `jobs.execute`)
- `QUEUE_DEAD_LETTER_SUBJECT` (default `jobs.dead_letter`)
- `QUEUE_CONSUMER_NAME` (default `jobs_worker`)
- `QUEUE_MAX_RETRIES` (default `3`)
- `QUEUE_BACKOFF_BASE` (default `1s`)

Exponential backoff uses:

- retry #1 => `QUEUE_BACKOFF_BASE`
- retry #2 => `2 * base`
- retry #3 => `4 * base`
- ...

## Reliability test evidence

`backend/internal/queue/nats_integration_test.go` runs failure injection against an embedded NATS server:

- Handler always fails.
- Message is delivered exactly `max_retries + 1` times.
- Message is dead-lettered after cap is reached.
- Test asserts no silent loss by observing a dead-letter event and attempt count.

## FastQC stage worker (Issue #10)

FastQC stage execution worker is implemented in `backend/internal/pipeline/fastqc`.

- Queue message includes `job_id` and payload (`input_path`, `output_dir`, optional report artifact URIs).
- Worker enforces per-job timeout via `context.WithTimeout` (`timeout_seconds` in payload or default config timeout).
- Worker updates job state to running/succeeded/failed and persists report artifact metadata in `output_ref`.
- Stage completion log emits `job_id`, `stage`, `exit_code`, and `duration`.
