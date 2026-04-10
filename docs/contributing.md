# Contributing Guide

This project follows a delivery standard focused on reproducibility, quality, and safe releases.

## Backend (Go)

Strict standards, testing, benchmarks, concurrency rules, linting, and library selection criteria are defined in **`docs/backend-engineering.md`**. All Go changes must comply with the **[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)** and pass **`golangci-lint`** (`backend/.golangci.yml`) and **`go test -race ./...`** in `backend/`.

## Branching and naming

- Create branches from `main`.
- Use one issue per branch.
- Branch naming:
  - `feature/<issue-number>-short-topic`
  - `task/<issue-number>-short-topic`
  - `bugfix/<issue-number>-short-topic`

## Pull request policy

- Link exactly one primary issue in each PR.
- Use the PR template and complete all sections.
- Include test evidence and rollback notes in every PR.
- Keep PR scope focused; avoid mixed-purpose changes.

## Required checks (merge gates)

The following checks are required before merge:

1. Lint passes.
2. Unit tests pass.
3. Integration tests pass (when applicable to changed modules).
4. OpenAPI/schema validation passes (for API changes).
5. Security scan has no unresolved critical findings.
6. Reviewer approval is present.

## Branch protection policy (target: `main`)

- Require pull request before merging.
- Require at least 1 approval.
- Require conversation resolution before merge.
- Require all required checks to pass.
- Dismiss stale approvals on new commits.
- Restrict direct pushes to `main`.

## Quality standards

- Every change must map to explicit acceptance criteria.
- Functional and non-functional requirements must both be verified.
- New behavior must include tests or documented rationale when tests are not practical.
- Documentation must be updated alongside behavior changes.

## Security and secrets

- Never commit secrets, credentials, or private keys.
- Use environment variables and local secret management for sensitive data.
- Report suspected secret exposure immediately and rotate affected credentials.

## Definition of done (project-wide)

A task is done only when all are true:

1. Acceptance criteria are met.
2. Required checks pass.
3. Test evidence is attached to PR.
4. Rollback approach is documented.
5. Docs are updated.
