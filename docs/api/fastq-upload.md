# FASTQ upload metadata API

Registers **paired-end FASTQ ingestion metadata** so the platform can create a `jobs` row before actual file transfer or pipeline execution. Implemented for [Issue #6](https://github.com/AminN77/senju/issues/6).

## Prerequisites

- API process configured with PostgreSQL (`POSTGRES_DSN` or `POSTGRES_HOST` + credentials).
- Schema applied: run migrations (see [docs/setup.md](../setup.md#database-migrations)).

## Endpoint

- **POST** `/v1/jobs/fastq-upload/metadata`
- **Content-Type:** `application/json`

### Request body (required fields)

| Field       | Description                                      |
|------------|--------------------------------------------------|
| `sample_id` | Sample identifier                               |
| `r1_uri`    | URI for read 1 (e.g. `https://...`, `s3://bucket/key/...`) |
| `r2_uri`    | URI for read 2                                 |

Optional: `library_id`, `platform`, `notes` (non-empty strings are stored in the job `input_ref` JSON).

### Success (201)

```json
{ "job_id": "550e8400-e29b-41d4-a716-446655440000" }
```

### Errors

- **400** `application/problem+json` — malformed JSON, unknown fields, or validation failures (`errors[]` sorted by `field`, then `message`).
- **500** `application/problem+json` — persistence failure when saving the job row.
- **503** `application/problem+json` — database not configured for the API process.

## Example

```bash
curl -sS -X POST "http://localhost:${API_PORT:-8080}/v1/jobs/fastq-upload/metadata" \
  -H "Content-Type: application/json" \
  -d '{
    "sample_id": "SAMPLE01",
    "r1_uri": "s3://my-bucket/run1/R1.fq.gz",
    "r2_uri": "s3://my-bucket/run1/R2.fq.gz",
    "library_id": "LIB-001"
  }'
```

OpenAPI: [backend/openapi/openapi.yaml](../../backend/openapi/openapi.yaml).
