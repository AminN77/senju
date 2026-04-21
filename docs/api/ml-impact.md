# ML impact baseline API

Issue: [#20](https://github.com/AminN77/senju/issues/20)

## Endpoints

- `POST /v1/ml/impact/train`
  - trains baseline impact model from labeled samples
  - returns model metadata (`dataset_hash`, `feature_version`, `model_version`)
- `POST /v1/ml/impact/{job_id}/predict`
  - predicts one variant score/class and persists result in `jobs.output_ref`

## Reproducibility metadata

Training response includes:

- `dataset_hash`: SHA-256 hash of canonicalized training samples
- `feature_version`: currently `impact_features_v1`
- `model_version`: generated training version identifier

Prediction persistence payload includes the same metadata for audit and replay.

## Latency target

Online inference target: p95 `<100ms` on representative input profile.

## Local verification

```bash
cd backend
go test ./internal/ml/impact ./internal/api/mlimpact ./internal/httpserver -count=1
```
