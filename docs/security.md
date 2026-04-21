# Security baseline

Issue: [#17](https://github.com/AminN77/senju/issues/17)

## API authentication and authorization

The API uses JWT bearer authentication with role-based authorization.

- Algorithm: `HS256`
- Required claims:
  - `iss` (issuer) must match `JWT_ISSUER`
  - `exp` (expiration) must be valid
  - `role` must be one of: `uploader`, `runner`, `analyst`, `admin`

### Protected routes

- Upload APIs (`uploader|admin`):
  - `POST /v1/jobs/fastq-upload/metadata`
  - `POST /v1/jobs/fastq-upload/{job_id}/validate`
  - `POST /v1/objects/multipart`
  - `POST /v1/objects/multipart/{upload_id}/parts`
  - `POST /v1/objects/multipart/{upload_id}/complete`
- Run/orchestration APIs (`runner|admin`):
  - `POST /v1/jobs/pipeline`
  - `POST /v1/jobs/{job_id}/run`
  - `GET /v1/jobs/{job_id}/status`
  - `GET /v1/jobs/{job_id}/outputs`
- Query APIs (`analyst|admin`):
  - `GET /v1/variants`

## Secret handling and rotation

### Required auth secrets

- `JWT_SECRET` (required): signing key for API JWT validation
- `JWT_ISSUER` (required): expected token issuer

### Rotation process

1. Generate a new strong secret in your secret manager (do not commit to git).
2. Update runtime environment (`JWT_SECRET`) in deployment platform.
3. Redeploy API instances with the new secret.
4. Invalidate and re-issue client tokens signed by the old key.
5. Verify `401` for old tokens and successful access with new tokens.
6. Record rotation timestamp and owner in your runbook/audit log.

For emergency rotation, follow the same flow immediately and shorten token TTL while rotating.

## CI security scanning

Workflow: `.github/workflows/security.yml`

- `govulncheck` for Go dependency/code vulnerabilities
- Trivy image scanning for container OS/library vulnerabilities
- PR/build fails on unresolved **CRITICAL** image findings
