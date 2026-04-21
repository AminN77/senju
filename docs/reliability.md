# Reliability controls

Issue: [#18](https://github.com/AminN77/senju/issues/18)

## Checkpoint/restart baseline

Current checkpoint/restart support is implemented for the alignment stage (`bwa` + `samtools`):

- checkpoint file: `<output_bam_path>.checkpoint.json`
- checkpoint is persisted only after `bwa mem` succeeds
- retries can resume from checkpoint and run `samtools sort` without re-running `bwa`

## Recovery objective assumptions (RTO)

For alignment stage interruptions where checkpoint exists:

- restart path skips expensive re-alignment and resumes sorting/output write
- expected RTO is bounded by:
  - queue redelivery delay (`QUEUE_BACKOFF_BASE`, retries)
  - worker startup latency
  - remaining `samtools sort` runtime

Without checkpoint (interruption before checkpoint persistence), the stage restarts from `bwa`.

## Chaos test evidence

`backend/internal/pipeline/alignment/worker_test.go` includes a chaos-style recovery test:

- first run simulates interruption after `bwa` and before successful `samtools`
- checkpoint is verified present
- second run resumes, skips `bwa`, runs `samtools`, and reaches `alignment_succeeded`
- output metadata includes checksum and checkpoint cleanup is verified
