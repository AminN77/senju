# Frontend Roadmap — UI Foundation Phase

This phase delivers a promotable, familiar-feeling UI so Senju can be shared publicly and evaluated against commercial genomics platforms. Every item listed maps to a GitHub issue.

The phase is organized into four sequenced tracks. Tracks 1 and 2 must be substantially complete before Track 3 ships user-visible screens; Track 4 runs in parallel.

**Milestone:** `M6 Frontend UI Foundation`
**Primary label:** `area:frontend`
**Phase label:** `phase:ui-foundation`

---

## Track 1 — Project foundation (tasks)

Goal: a buildable, lintable, testable, deployable Next.js app with CI, before any screen work begins.

| # | Title | Type | Depends on |
| --- | --- | --- | --- |
| F1 | Scaffold `frontend/` Next.js + TypeScript + pnpm app | Task | — |
| F2 | Configure Tailwind + CSS-variable token pipeline | Task | F1 |
| F3 | Install and configure shadcn/ui (`components.json`, generators) | Task | F2 |
| F4 | Set up Vitest + RTL + jest-axe + Playwright | Task | F1 |
| F5 | Set up Storybook 8 with a11y + interactions addons | Task | F3 |
| F6 | Add frontend CI workflow (lint, typecheck, test, a11y, build, bundle budget) | Task | F4, F5 |
| F7 | Add pre-commit hooks (lint-staged, typecheck, test:changed) | Task | F4 |
| F8 | Configure OpenAPI codegen to typed TS client | Task | F1 |
| F9 | Implement authentication client + JWT session handling | Task | F8 |
| F10 | Implement global app shell (sidebar, topbar, theme toggle, breadcrumb, skip link) | Task | F2, F3 |

---

## Track 2 — Design system (features)

Goal: the token system, palette, core components, and domain components that every screen relies on.

| # | Title | Type | Depends on |
| --- | --- | --- | --- |
| DS1 | Establish design token system (colors, spacing, typography, radii, elevation, motion, z-index) | Feature | F2 |
| DS2 | Implement color palette — dark + light themes with WCAG AA contrast test | Feature | DS1 |
| DS3 | Typography scale and font loading (Inter + JetBrains Mono via `next/font`) | Task | DS1 |
| DS4 | Core components batch 1 — Button, Input, Select, Checkbox, Radio, Switch, Label, FormField | Feature | DS1, F3 |
| DS5 | Core components batch 2 — Dialog, Dropdown, Popover, Tooltip, Tabs, Toast, Banner, Skeleton | Feature | DS4 |
| DS6 | Data display components — DataTable (virtualized), Pagination, FilterBar, EmptyState, ErrorState | Feature | DS5 |
| DS7 | Domain components — JobStatusPill, PipelineStageTimeline, VariantRow, FileUploadDropzone, LogViewer | Feature | DS6 |
| DS8 | Accessibility standards — axe-core in CI, keyboard walkthrough checklist, contrast test | Task | F6, DS2 |

---

## Track 3 — Screens (features)

Goal: every screen in the phase scope, built on Track 1 + Track 2.

| # | Title | Type | Depends on |
| --- | --- | --- | --- |
| S1 | Marketing landing page (`/`) | Feature | DS5 |
| S2 | Authentication screens (`/login`, `/signup`, `/forgot-password`, `/reset-password`, `/verify-email`) | Feature | DS5, F9 |
| S3 | Workspace/project dashboard (`/dashboard`) | Feature | DS7, F10 |
| S4 | FASTQ upload flow (`/projects/:id/upload`) | Feature | DS7 |
| S5 | Pipeline jobs list (`/jobs`) | Feature | DS6, DS7 |
| S6 | Pipeline job detail (`/jobs/:id`, stages, logs, artifacts) | Feature | DS7 |
| S7 | Variant browser (`/variants`, `/variants/:id`) | Feature | DS6, DS7 |
| S8 | ML variant-impact scoring UI (`/ml/impact`) | Feature | DS7 |
| S9 | Admin console (users, roles, tokens, audit) | Feature | DS6 |
| S10 | Account settings (profile, security, API keys, preferences) | Feature | DS5 |

---

## Track 4 — Cross-cutting (parallel)

- Content and copywriting review for marketing + empty states + error messages.
- Contributor docs (`frontend/README.md`, `frontend/CONTRIBUTING.md`).
- Analytics decision ADR (opt-in, privacy-respecting — deferred to a separate issue if in-scope).
- License audit (`frontend/LICENSES.md` auto-generated).

---

## Exit criteria for the phase

All of the following must be true:

1. All Track 1, Track 2, and Track 3 issues closed.
2. WCAG 2.1 AA axe-core checks pass on every shipped route in CI.
3. Contrast test passes for both themes, all production token pairs.
4. Bundle-size budgets met per `design-principles.md`.
5. Storybook published (internal or Chromatic — decision in the Storybook setup issue).
6. End-to-end Playwright smoke of primary auth + dashboard + jobs list + variant browser + upload passes.
7. Docs (`docs/frontend/*`) up to date with any drift from shipped implementation.
8. README updated to link the frontend and its docs.
9. IP review (see `ip-and-references.md` §3) signed off by two reviewers.

When met, the phase closes and a post-phase retrospective ADR captures what to change for the next phase.

---

## What this phase explicitly does NOT include

- Mobile apps.
- Billing / subscription surfaces.
- Collaboration / commenting / sharing workflows.
- Real-time multi-user presence.
- Rich reporting / PDF export.
- Plugin/marketplace surface.
- Any SSO integration beyond a UI placeholder (backend integration is a separate phase).
- i18n translations (scaffolding is also out; English-only by explicit decision).

Every one of these is a legitimate future phase; none of them gate promotion of the current UI.
