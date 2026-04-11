# Object multipart upload API (S3-compatible)

Implements presigned **multipart uploads** for large objects (including multi‑GB files via many parts) against **S3-compatible** storage such as **MinIO** ([Issue #7](https://github.com/AminN77/senju/issues/7)).

## Prerequisites

Configure the API process with:

| Variable | Description |
|----------|-------------|
| `S3_ENDPOINT` | Base URL of the S3 API (e.g. `http://localhost:9001` from the host, or `http://minio:9000` inside Compose). |
| `S3_BUCKET` | Target bucket name (create it once, e.g. with `mc mb` or MinIO Console). |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | API credentials. If unset, `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` are used. |
| `S3_REGION` | AWS signing region (default `us-east-1`; required by the SDK even for MinIO). |
| `S3_USE_PATH_STYLE` | `true` for MinIO with a custom endpoint (default `true`). |

If any of endpoint, bucket, access key, or secret key are missing, **multipart routes return 503**.

`docker-compose.yml` sets `S3_ENDPOINT` and `S3_BUCKET` for the API service; ensure the bucket exists (for example `mc mb local/senju-uploads` against your MinIO alias).

## Flow

1. **POST** `/v1/objects/multipart` — start upload; receive `object_key`, `upload_id`, `part_size_bytes` (64 MiB).
2. For each part: **POST** `/v1/objects/multipart/{upload_id}/parts` with `object_key` and `part_number` — receive a presigned **PUT** `url`.
3. **PUT** part bytes to `url` (directly to MinIO). Record the **ETag** response header.
4. **POST** `/v1/objects/multipart/{upload_id}/complete` with `object_key` and `parts` (part numbers and ETags). The server completes the multipart upload, runs **HeadObject** to verify size/ETag, and writes an **audit log** line (`checksum_status=validated`).

Part size: use the returned `part_size_bytes` for all parts except the last; the last part may be smaller. For a single-part object, one small part is valid.

## Endpoints

- **POST** `/v1/objects/multipart` — `Content-Type: application/json`  
  Body: `{ "content_type": "<mime>", "filename_hint": "<optional>" }`  
  **201:** `{ "bucket", "object_key", "upload_id", "part_size_bytes" }`

- **POST** `/v1/objects/multipart/{upload_id}/parts`  
  Body: `{ "object_key": "...", "part_number": 1 }`  
  **200:** `{ "url", "expires_at" }`

- **POST** `/v1/objects/multipart/{upload_id}/complete`  
  Body: `{ "object_key": "...", "parts": [ { "part_number": 1, "etag": "..." } ] }`  
  **200:** `{ "object_key", "etag", "size_bytes" }`

## Errors

- **400** `application/problem+json` — validation / malformed JSON.
- **500** — object store operation failed.
- **503** — object storage not configured.

OpenAPI: [backend/openapi/openapi.yaml](../../backend/openapi/openapi.yaml).

## Integration test (optional)

With MinIO listening (e.g. Compose), create the bucket, then:

```bash
cd backend
RUN_MINIO_INTEGRATION=1 S3_ENDPOINT=http://localhost:9001 S3_BUCKET=senju-uploads \
  MINIO_ROOT_USER=minioadmin MINIO_ROOT_PASSWORD=minioadmin \
  go test ./internal/objectstore/ -run TestS3Store_MultipartRoundTrip -count=1 -v
```
