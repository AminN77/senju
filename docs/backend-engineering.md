# Backend engineering standards (Go)

This document is the **authoritative bar** for Senju backend work: maintainability, idiomatic Go, performance discipline, tests, and automation. It complements [`contributing.md`](contributing.md).

## Normative references (read and follow)

1. **[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)** — **mandatory** for style, errors, concurrency, and patterns unless this doc explicitly overrides.
2. [Effective Go](https://go.dev/doc/effective_go)
3. [Go Wiki: Common Mistakes](https://go.dev/wiki/CommonMistakes)
4. [Go Wiki: Code Review Comments](https://go.dev/wiki/CodeReviewComments)

If advice conflicts, order of precedence is: **this doc for Senju-specific rules** → **Uber guide** → Effective Go / wiki.

## Principles

- **Maintainability first:** clear boundaries, small packages, explicit APIs, and documentation where behavior is non-obvious.
- **Performance:** optimize **measured** hot paths; avoid premature optimization elsewhere. Prefer algorithms and data structures that scale; avoid unnecessary allocations in loops.
- **Correct concurrency:** safe shutdown, no leaked goroutines, context propagation, and clear ownership of channels and mutexes.
- **Evidence:** changes to performance-sensitive code include **benchmarks** and/or **profiles** (see Testing).

## Library and dependency selection

When choosing a third-party library (HTTP routers, DB drivers, serialization, etc.):

1. **Prefer mature, actively maintained** modules with a clear license compatible with the project.
2. **Compare performance** using realistic workloads: read upstream benchmarks, community comparisons, and/or add **local `Benchmark*`** tests before committing to a dependency for hot paths.
3. **Prefer the best-in-class or among the best** for the intended use case when trade-offs are otherwise equal (API fit, security, maintenance).
4. **Minimize dependency weight:** avoid pulling large transitive trees for a small feature; prefer stdlib or thin wrappers when performance and clarity are comparable.
5. **Pin versions** in `go.mod` / `go.sum`; document non-obvious choices in an ADR or package doc comment.

## Architecture and design patterns

- **Boundaries:** define **interfaces at package boundaries** (consumers depend on interfaces; implementations live behind constructors). Align with ADRs for queues, DBs, and workflows.
- **Dependency injection:** prefer explicit constructor injection over global singletons (Uber: avoid mutable globals).
- **Options pattern:** use **functional options** for optional configuration on constructors where it improves clarity (Uber Patterns section).
- **Layers:** keep HTTP handlers thin; put domain logic in testable packages; isolate I/O (DB, object storage, messaging) behind interfaces.
- **Errors:** wrap with `%w` where appropriate; handle errors once; no silent ignores (enforce via lint + review).

## Concurrency and goroutines (strict)

Follow the Uber guide sections on goroutines and apply these Senju rules:

1. **No fire-and-forget goroutines** without lifecycle control (cancellation, timeout, or tracked completion).
2. **Always wait** for goroutines to finish on shutdown (e.g. `sync.WaitGroup`, `errgroup`, or parent `context` cancellation with joined workers).
3. **Propagate `context.Context`** for cancellation and deadlines on I/O and long work.
4. **Channel sizing:** default to **size 0 or 1** unless a bounded buffer is explicitly justified (Uber: *Channel Size is One or None*).
5. **Synchronization:** prefer `sync.Mutex` / `RWMutex` for shared memory; document invariants; use `go.uber.org/atomic` or `sync/atomic` for simple counters when appropriate (per Uber).
6. **No goroutines in `init()`** (Uber).
7. **Tests:** use **`-race`** for concurrent code paths in CI where practical.

## Performance practice

- Prefer **`strconv`** over **`fmt`** for hot numeric/string conversion (Uber Performance).
- **Preallocate** slices/maps when size is known or bounded.
- Avoid repeated **string ↔ []byte** conversions in tight loops.
- **Profile** (`pprof` CPU/heap) before large refactors claimed as “performance” work.
- Document complexity (time/space) for non-trivial algorithms in exported functions.

## Testing and benchmarks

| Area | Requirement |
|------|-------------|
| **Unit tests** | Table-driven tests where multiple cases exist (Uber). Cover edge cases and failure modes. |
| **Integration** | Use for DB, queue, and object storage boundaries; gate with build tags or env if needed. |
| **Race** | `go test -race ./...` in CI for `backend` (required job). |
| **Benchmarks** | Add `Benchmark*` for **hot paths** and **allocation-sensitive** code (handlers on critical paths, parsers, serializers, batch processors). |
| **Coverage** | No fixed global % in MVP; **new packages must not decrease** coverage of critical logic without justification in PR. Aim upward over time. |

Benchmark naming: `BenchmarkFoo` in `*_test.go` next to implementation.

## Linting and formatting

- **Formatting:** `gofmt` / `goimports` (or `golangci-lint` formatters) — no manual style debates in review.
- **Linting:** **`golangci-lint`** using [`backend/.golangci.yml`](../backend/.golangci.yml) runs locally and in CI.
- **Vet:** `go vet ./...` remains a required check.

Local quick checks from repo root:

```bash
cd backend && golangci-lint run ./...
cd backend && go test -race ./...
```

Install: see [golangci-lint install](https://golangci-lint.run/welcome/install/).

## Code review checklist (backend)

- [ ] Complies with **Uber Go Style Guide** and this document.
- [ ] Interfaces and package boundaries are clear; no unnecessary exported symbols.
- [ ] Errors handled; no ignored `err` without `//nolint` + comment.
- [ ] Goroutines have a clear lifecycle; shutdown path tested where applicable.
- [ ] Tests added/updated; **benchmarks** updated for touched hot paths.
- [ ] No new dependencies without benchmark or maintenance justification in PR description.
- [ ] OpenAPI / contracts updated when HTTP API changes.

## Evolution

When standards change, update **this file** and, if needed, **ADRs**. CI and `backend/.golangci.yml` should stay in sync with the rules here.
