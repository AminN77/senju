# Variant storage and ingestion (Issue 13)

This document defines the ClickHouse variant schema and the VCF ingestion loader contract.

## ClickHouse table

The loader ensures the `variants` table exists:

- Database: current connection database (default `senju` in local setup)
- Table: `variants`
- Engine: `ReplacingMergeTree`
- Sort key: `(dataset_id, chrom, pos, ref, alt)`

Columns:

- `dataset_id String`
- `chrom LowCardinality(String)`
- `pos UInt32`
- `ref String`
- `alt String`
- `qual Nullable(Float64)`
- `filter String`
- `info String`
- `source_key String`

`source_key` is a deterministic hash of `(dataset_id, chrom, pos, ref, alt)` used for duplicate protection and idempotent reruns.

## Loader behavior

Package: `backend/internal/variant/clickhouse`

- Accepts a VCF stream (`io.Reader`) and parses non-header rows.
- Splits multiallelic ALT values (`A,C`) into separate inserted variants.
- Inserts with duplicate protection:
  - In-run duplicates are skipped via an in-memory seen-set.
  - Re-run duplicates are skipped by batched key lookups (`SELECT source_key ... IN (...)`) against `(dataset_id, source_key)` before batched inserts.
- Returns inserted row count for observability and test assertions.

## Running tests and benchmark

From `backend/`:

```bash
go test ./internal/variant/clickhouse -run TestParseLine
go test ./internal/variant/clickhouse -run TestLoadVCF_IdempotentIntegration
go test ./internal/variant/clickhouse -bench BenchmarkLoadVCF_ParseOnlyThroughput -benchmem
```

Integration test requires `CLICKHOUSE_DSN`. Example:

```bash
export CLICKHOUSE_DSN='clickhouse://default:senju_dev_password@localhost:9000/senju'
```
