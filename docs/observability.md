# Observability baseline

Issue: [#16](https://github.com/AminN77/senju/issues/16)

This baseline provides:

- Prometheus metrics collection
- centralized logs via Loki (ingested by Promtail)
- Grafana with pre-provisioned data sources and dashboard

## Stack (Docker Compose)

- Prometheus: `http://localhost:9090`
- Loki: `http://localhost:3100`
- Grafana: `http://localhost:3000`
- Promtail: ships container logs to Loki

## Pipeline stage metrics

All pipeline stages emit:

- `senju_pipeline_stage_duration_seconds` (histogram, labels: `stage`, `outcome`)
- `senju_pipeline_stage_total` (counter, labels: `stage`, `outcome`, `error_class`)

Stages covered:

- `fastqc`
- `alignment`
- `gatk`

## Grafana dashboard

Provisioned dashboard UID: `senju-pipeline-observability`

Panels include:

- p95 stage latency
- stage error rate
- SLO compliance panel (99% success over 30d)
- error budget burn panel (1h window vs 99% SLO)
- centralized logs panel from Loki

## Quick verification

```bash
docker compose up -d
curl http://localhost:9090/api/v1/targets
curl http://localhost:8080/metrics | rg "senju_pipeline_stage_"
```
