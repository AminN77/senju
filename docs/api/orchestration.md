# Pipeline orchestration API

Issue: [#15](https://github.com/AminN77/senju/issues/15)

This API provides end-to-end orchestration flow for pipeline jobs:

- create job
- run pipeline orchestration
- fetch status
- retrieve outputs (including transition audit metadata)

Base path: `/v1/jobs`

## Endpoints

### `POST /pipeline`

Creates a pipeline orchestration job in `pending` / `pipeline_queued`.

Request body:

```json
{
  "sample_id": "S-001",
  "r1_uri": "s3://bucket/sample_R1.fastq.gz",
  "r2_uri": "s3://bucket/sample_R2.fastq.gz",
  "force_fail": false
}
```

`force_fail` is an optional controlled-failure hook used to validate failure behavior in tests.

Response `201`:

```json
{ "job_id": "1f9f8e43-2fd2-4988-89e0-4c9b2b8a2148" }
```

### `POST /{job_id}/run`

Runs orchestration for a queued job.

- Transitions: `pipeline_queued -> pipeline_running -> pipeline_succeeded|pipeline_failed`
- Transition metadata is persisted in `output_ref.transition_log`.
- Returns `409` if the job is not runnable.

Response `200`:

```json
{
  "job_id": "1f9f8e43-2fd2-4988-89e0-4c9b2b8a2148",
  "status": "succeeded",
  "stage": "pipeline_succeeded"
}
```

### `GET /{job_id}/status`

Returns status, stage, and timestamps.

### `GET /{job_id}/outputs`

Returns orchestration output payload (`output_ref`) captured on the job row.

## Local verification

```bash
cd backend
go test ./internal/api/orchestration -count=1
go test ./... -count=1
```
