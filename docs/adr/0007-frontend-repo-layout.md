# ADR-0007: Frontend repo layout (monorepo `frontend/` subtree)

- **Status:** Accepted
- **Date:** 2026-04-21
- **Owners:** Platform team
- **Decision scope:** Where the frontend code lives relative to the existing `backend/` Go project and supporting governance.

## Context

The repository currently contains `backend/` (Go), `docs/`, `deploy/`, `proposals/`, and top-level tooling. The frontend is greenfield. Options range from a separate repo to a polyrepo with packages.

Project constraints:

- Small team; low operational overhead.
- Backend and frontend share an API contract (`backend/openapi/openapi.yaml`) — colocating makes the contract discoverable and changes reviewable in one PR.
- Existing Cursor rules and PR template work across the whole repo; fragmenting would duplicate governance.

## Decision

1. **Layout:** A top-level **`frontend/`** directory in this repository, mirroring the existing `backend/` pattern.
2. **Tooling boundary:** Frontend and backend have independent tooling — separate `package.json` (frontend), separate Go module (backend), separate lint/test commands, separate CI jobs.
3. **Shared contract:** The frontend generates a typed API client from `backend/openapi/openapi.yaml` at build time. No copy-paste of types across boundaries.
4. **No workspace tooling yet:** We do **not** adopt pnpm workspaces / Turborepo / Nx at this phase. If we later extract `packages/ui` or `packages/design-tokens`, we'll introduce pnpm workspaces in a new ADR.
5. **CI:** A separate GitHub Actions workflow (`.github/workflows/frontend.yml`) runs on changes under `frontend/**` and `backend/openapi/**`.

## Alternatives considered

1. **Separate repo (`senju-web`)**
   - Pros: independent release cadence.
   - Cons: API-contract drift, duplicated governance (rules, templates, CODEOWNERS), cross-repo PR coordination, higher contributor friction.
2. **`apps/` + `packages/` pnpm workspace now**
   - Pros: ready for multi-app scale.
   - Cons: premature — we have one frontend app and no shared packages yet. Adopt when justified.
3. **Embed frontend inside the Go binary (`go:embed`)**
   - Pros: single-binary deployment.
   - Cons: couples frontend build to backend release; harder to iterate on UI; against Next.js's rendering model.

## Trade-offs

- **Gain:** One repo, one governance model, one issue tracker; API contract changes and consumer changes are reviewable together.
- **Cost:** CI must be scoped by path filters to avoid running backend tests for frontend-only PRs and vice versa.
- **Constraint:** Frontend and backend must remain decoupled at the deployment level (separate containers, separate Dockerfiles).

## Risks

- **Blurred ownership** if contributors modify both trees in one PR — mitigated by the existing PR scope rule (`senju-workflow`: "stay scoped, no mixed unrelated changes") and CODEOWNERS updates when frontend owners exist.

## Migration path

- If the frontend outgrows the monorepo (e.g., independent release cadence becomes critical), `frontend/` is self-contained and can be extracted with `git subtree` / `git filter-repo` without touching backend code.

## Consequences

`frontend/` is a self-contained Next.js app alongside `backend/`. Governance (rules, templates, PR workflow) applies uniformly. CI is path-scoped.
