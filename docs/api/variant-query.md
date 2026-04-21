# Variant query API (Issue 14)

`GET /v1/variants` queries ClickHouse variant rows with safe filters and pagination.

## Query parameters

- `chromosome` (optional): exact chromosome match (for example `chr1`)
- `position_min` (optional): inclusive lower bound
- `position_max` (optional): inclusive upper bound
- `gene` (optional): gene symbol filter (`GENE=` token in INFO)
- `page` (optional, default `1`, min `1`)
- `page_size` (optional, default `50`, min `1`, max `200`)

## Validation rules

- `chromosome` and `gene` must match `^[A-Za-z0-9._-]+$`.
- `position_min`/`position_max` must be unsigned 32-bit integers.
- if both position bounds are provided, `position_min <= position_max`.
- invalid inputs return RFC7807 `400 application/problem+json`.

## Response shape

`200 application/json` returns:

- `data`: list of variants (`chromosome`, `position`, `ref`, `alt`, `qual`, `filter`, `info`, `gene`)
- pagination metadata: `page`, `page_size`, `total`, `has_next`
- observability fields: `latency_ms`, applied `filters`

## Local verification

From `backend/`:

```bash
go test ./internal/api/variantquery -count=1
CLICKHOUSE_DSN='clickhouse://default:senju_dev_password@localhost:9000/senju' go test ./internal/variant/clickhouse -run 'TestQueryRepository_IntegrationFiltersAndPagination|TestQueryRepository_P95Under500ms' -count=1
```
