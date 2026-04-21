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

## Alignment stage worker (Issue #11)

BWA + SAMtools alignment worker is implemented in `backend/internal/pipeline/alignment`.

- Queue message payload includes `reference_path`, `read1_path`, `read2_path`, and `output_bam_path` (optional `output_bam_uri`).
- Worker executes `bwa mem` then `samtools sort`, and persists BAM artifact location + checksum (`checksum_sha256`) in `output_ref`.
- Stage-level CPU/memory limits are configurable per job (`threads`, `memory_limit_mb`) with safe defaults from worker config.
- Deterministic output verification is documented and tested: same input/config and output bytes produce same SHA-256 checksum.
- Stage logs include `job_id`, `stage=alignment`, `exit_code`, and `duration`.

## GATK stage worker (Issue #12)

GATK variant-calling worker is implemented in `backend/internal/pipeline/gatk`.

- Payload fields: `reference_path`, `input_bam_path`, `output_vcf_path` (+ optional `output_vcf_uri`).
- Worker executes `gatk HaplotypeCaller`, persists VCF artifact metadata in `output_ref`, and updates status transitions (`gatk_running` -> `gatk_succeeded` / `gatk_failed`).
- Errors are classified into `tool`, `infrastructure`, and `configuration`.
- Stage metrics are exported to Prometheus:
  - `senju_pipeline_stage_duration_seconds` (histogram)
  - `senju_pipeline_stage_total` (counter with `stage/outcome/error_class` labels)

## Pipeline stage observability baseline (Issue #16)

All pipeline stages (`fastqc`, `alignment`, `gatk`) now emit the same Prometheus metric families:

- `senju_pipeline_stage_duration_seconds` (histogram)
- `senju_pipeline_stage_total` (counter)

Labels are consistent across stages:

- `stage`
- `outcome` (`success`/`failure`)
- `error_class` (empty for success; failure classification when available)
