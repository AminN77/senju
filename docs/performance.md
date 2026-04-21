# Performance qualification and regression gates

Issue: [#19](https://github.com/AminN77/senju/issues/19)

## Benchmark coverage

The repository includes targeted benchmarks for core performance paths:

- Upload throughput:
  - `BenchmarkMultipartInit_Throughput` (`internal/api/objectupload`)
- Pipeline throughput:
  - `BenchmarkAlignmentHandle_Throughput` (`internal/pipeline/alignment`)
- Query latency:
  - `BenchmarkGetVariants_Latency` (`internal/api/variantquery`)

## Baseline targets and thresholds

Stored in repository:

- baseline file: `backend/perf/baseline.json`
- comparator tool: `backend/cmd/perfcheck`

`baseline.json` defines:

- per-benchmark `target_ns_op`
- warn threshold (`warn_regression_pct`)
- fail threshold (`fail_regression_pct`)
- targets are calibrated for CI runner variability, while still detecting meaningful regressions

## CI regression policy

`Backend CI` runs a performance gate step:

1. execute benchmark subset
2. compare measured `ns/op` against baseline targets
3. emit:
   - `PASS` when within warn threshold
   - `WARN` when regression > warn threshold
   - `FAIL` when regression > fail threshold (CI failure)

## Local verification

```bash
cd backend
go test -run '^$' -bench 'BenchmarkMultipartInit_Throughput|BenchmarkAlignmentHandle_Throughput|BenchmarkGetVariants_Latency' -benchmem -count=1 ./internal/api/objectupload ./internal/pipeline/alignment ./internal/api/variantquery > /tmp/bench.txt
go run ./cmd/perfcheck
```
