# Frontend Engineering Docs

This folder defines standards and decisions for the Senju web frontend. It mirrors the role `docs/backend-engineering.md` plays for the Go backend.

Read this page first if you are contributing to anything under `frontend/`.

## Scope of this phase (UI Foundation)

Goal: ship a promotable, familiar-feeling UI surface so Senju can be shared publicly and evaluated by users migrating from commercial platforms (primary UX reference: DNAnexus — patterns only, visual identity is original; see `ip-and-references.md`).

In-phase screens:

- Marketing landing page
- Authentication (login, signup, password reset, verification)
- Workspace/project dashboard
- FASTQ upload flow (multipart presigned)
- Pipeline jobs list + filters
- Pipeline job detail (stages, logs, metrics, artifacts)
- Variant browser (filters, pagination, row detail)
- ML variant-impact scoring UI
- Admin console (users, RBAC, tokens, audit)
- Account settings

Explicitly out of this phase: mobile apps, reporting/PDF export, collaboration/commenting, billing/subscription, data-room sharing, cross-workspace notifications.

## Documents

| Document | Purpose |
| --- | --- |
| [`design-principles.md`](./design-principles.md) | North-star principles that every design/engineering decision rolls up to |
| [`design-system.md`](./design-system.md) | Tokens, typography, spacing, radii, elevation, motion, component catalog |
| [`color-palette.md`](./color-palette.md) | Brand, neutrals, semantic palettes, data-viz scale, dark/light theme mapping, WCAG AA contrast table |
| [`information-architecture.md`](./information-architecture.md) | Route map, navigation shape, page inventory |
| [`accessibility.md`](./accessibility.md) | WCAG 2.1 AA baseline, keyboard, focus, semantics, testing |
| [`component-standards.md`](./component-standards.md) | Component authoring rules: file layout, tests, stories, a11y, API shape |
| [`ip-and-references.md`](./ip-and-references.md) | What is OK and NOT OK when using DNAnexus or other competitors as UX references |
| [`roadmap.md`](./roadmap.md) | Sequenced delivery plan mapping to GitHub issues |

## Related ADRs

- [ADR-0005: Frontend framework (Next.js + TypeScript)](../adr/0005-frontend-framework.md)
- [ADR-0006: Frontend design system (shadcn/ui + Radix + Tailwind + CSS-variable tokens)](../adr/0006-frontend-design-system.md)
- [ADR-0007: Frontend repo layout (monorepo `frontend/` subtree)](../adr/0007-frontend-repo-layout.md)

## Related Cursor rules

- `.cursor/rules/frontend-web.mdc` — engineering standards for `frontend/**`
- `.cursor/rules/frontend-design-system.mdc` — token discipline and design-system usage
- `.cursor/rules/frontend-ip-guardrails.mdc` — IP safety guardrails for the UI Foundation phase

## Governance

The frontend inherits the project's PR policy, issue templates, and merge gates from `docs/contributing.md` and `.cursor/rules/senju-workflow.mdc`. Frontend-specific additions:

- `area:frontend` label on every frontend issue/PR.
- Path-scoped CI workflow (`.github/workflows/frontend.yml`) runs lint, typecheck, unit tests, a11y tests, and bundle-size checks.
- No frontend PR merges without Storybook stories for new components and axe-core passing checks for affected routes.
