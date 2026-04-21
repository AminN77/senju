# Design System

This document specifies the token system, typography, spacing, radii, elevation, motion, and component catalog for the Senju frontend.

The design system is **token-driven**. Every visual value is a named token exposed as a CSS custom property and consumed by Tailwind utilities. Themes are token remappings, not component rewrites.

Color tokens are specified separately in [`color-palette.md`](./color-palette.md).

---

## 1. Token categories

| Category | Prefix | Examples |
| --- | --- | --- |
| Color — surface | `--color-surface-*` | `surface-base`, `surface-raised`, `surface-sunken`, `surface-overlay` |
| Color — text | `--color-text-*` | `text-primary`, `text-secondary`, `text-muted`, `text-on-accent` |
| Color — border | `--color-border-*` | `border-subtle`, `border-default`, `border-strong`, `border-focus` |
| Color — brand | `--color-brand-*` | `brand-50` … `brand-900` |
| Color — semantic | `--color-{success,warning,danger,info}-*` | `success-solid`, `warning-subtle`, etc. |
| Color — data viz | `--color-dv-*` | `dv-1` … `dv-10` (categorical), `dv-seq-*`, `dv-div-*` |
| Spacing | `--space-*` | `space-0`, `space-1` (4px) … `space-24` (96px) |
| Radius | `--radius-*` | `radius-xs`, `radius-sm`, `radius-md`, `radius-lg`, `radius-full` |
| Elevation | `--shadow-*` | `shadow-0`, `shadow-1`, `shadow-2`, `shadow-3`, `shadow-overlay` |
| Typography | `--font-*`, `--text-*`, `--leading-*`, `--tracking-*` | `font-sans`, `text-body-md`, `leading-body` |
| Motion | `--duration-*`, `--ease-*` | `duration-fast`, `ease-out-quad` |
| Z-index | `--z-*` | `z-base`, `z-dropdown`, `z-modal`, `z-toast` |

All tokens are defined in `frontend/src/styles/tokens.css` (authoritative) and surfaced through `tailwind.config.ts` so `bg-surface-raised`, `text-text-primary`, `border-border-default`, etc. work as Tailwind utilities.

---

## 2. Typography scale

**Fonts**

- Sans (UI): **Inter** — variable font, subset Latin + Latin-Ext, loaded via `next/font/google` with `display: swap`.
- Mono (code, logs, IDs): **JetBrains Mono** — variable font, loaded via `next/font/google`.

**Scale** (font-size / line-height / letter-spacing):

| Token | Size | Line height | Tracking | Use |
| --- | --- | --- | --- | --- |
| `text-display-lg` | 48px | 56px | -0.02em | Marketing hero only |
| `text-display-md` | 36px | 44px | -0.02em | Marketing section titles |
| `text-heading-xl` | 28px | 36px | -0.01em | Page titles (app shell) |
| `text-heading-lg` | 22px | 30px | -0.01em | Section titles |
| `text-heading-md` | 18px | 26px | 0 | Card titles, dialog titles |
| `text-heading-sm` | 16px | 24px | 0 | Subsection titles |
| `text-body-lg` | 16px | 26px | 0 | Default marketing body |
| `text-body-md` | 14px | 22px | 0 | **Default app body** |
| `text-body-sm` | 13px | 20px | 0 | Secondary info, metadata |
| `text-caption` | 12px | 18px | 0.01em | Labels, badges, captions |
| `text-code-md` | 13px | 20px | 0 | Inline code, IDs |
| `text-code-sm` | 12px | 18px | 0 | Log lines, dense mono |

Weights: 400 regular, 500 medium, 600 semibold, 700 bold. No 300 or 800.

---

## 3. Spacing scale

4px base. Use named steps, not arbitrary values.

| Token | Value | Typical use |
| --- | --- | --- |
| `space-0` | 0 | — |
| `space-1` | 4px | Icon-text gap |
| `space-2` | 8px | Compact inline gap |
| `space-3` | 12px | Button padding Y |
| `space-4` | 16px | Default gap |
| `space-5` | 20px | Card padding (compact) |
| `space-6` | 24px | Card padding (default) |
| `space-8` | 32px | Section gap |
| `space-10` | 40px | Large section gap |
| `space-12` | 48px | Page padding Y (app) |
| `space-16` | 64px | Marketing section gap |
| `space-20` | 80px | — |
| `space-24` | 96px | Marketing hero padding Y |

---

## 4. Radii

| Token | Value | Use |
| --- | --- | --- |
| `radius-xs` | 2px | Pills, tags |
| `radius-sm` | 4px | Inputs, small buttons |
| `radius-md` | 6px | **Default** — buttons, cards |
| `radius-lg` | 10px | Dialogs, popovers |
| `radius-xl` | 16px | Large surfaces, marketing cards |
| `radius-full` | 9999px | Avatars, status dots |

---

## 5. Elevation (shadows)

Dark and light themes have separate shadow definitions. Dark theme uses layered `rgba(0,0,0,…)` with a thin top highlight; light theme uses softer diffusion.

| Token | Use |
| --- | --- |
| `shadow-0` | No elevation (flat) |
| `shadow-1` | Resting cards, dropdowns |
| `shadow-2` | Hovered cards, menus |
| `shadow-3` | Popovers |
| `shadow-overlay` | Dialogs, modals |

Elevation is additive with `surface-*` tokens — a "raised" card uses `bg-surface-raised` + `shadow-1`, not a custom shadow.

---

## 6. Motion

| Token | Value | Use |
| --- | --- | --- |
| `duration-instant` | 80ms | Hover state changes |
| `duration-fast` | 150ms | Dropdowns, tooltips |
| `duration-base` | 200ms | **Default** — most transitions |
| `duration-slow` | 300ms | Dialog open/close |
| `duration-page` | 400ms | Route-level transitions |

| Easing token | Curve | Use |
| --- | --- | --- |
| `ease-out-quad` | cubic-bezier(0.25, 0.46, 0.45, 0.94) | Enter animations |
| `ease-in-out-quad` | cubic-bezier(0.455, 0.03, 0.515, 0.955) | Position changes |
| `ease-spring` | cubic-bezier(0.34, 1.56, 0.64, 1) | Playful accents only |

**Prefers-reduced-motion:** every animation must have a reduced-motion fallback that collapses to a 0ms state change or an opacity-only transition.

---

## 7. Z-index layers

| Token | Value | Use |
| --- | --- | --- |
| `z-base` | 0 | Page content |
| `z-sticky` | 10 | Sticky headers within pages |
| `z-topbar` | 20 | App top bar |
| `z-dropdown` | 30 | Dropdowns, popovers |
| `z-modal-backdrop` | 40 | Dialog/modal backdrop |
| `z-modal` | 41 | Dialog/modal content |
| `z-toast` | 50 | Toasts, notifications |
| `z-tooltip` | 60 | Tooltips (always on top) |

No custom z-index values — use tokens only.

---

## 8. Breakpoints

Tailwind default breakpoints, with opinion:

| Token | Min width | Target |
| --- | --- | --- |
| `sm` | 640px | Small tablet |
| `md` | 768px | Tablet |
| `lg` | 1024px | **Desktop baseline** |
| `xl` | 1280px | Wide desktop |
| `2xl` | 1536px | Ultrawide |

The authenticated console targets **`lg` and above** as primary. Mobile is responsive-degraded, not feature-complete (explicit out-of-phase decision).

---

## 9. Component catalog

Components are grouped into three tiers. Each tier has escalating review requirements (see `component-standards.md`).

### Tier 1 — Primitives (shadcn/ui + Radix)

These are owned in `frontend/src/components/ui/` (shadcn generates source files into our repo).

- Button, IconButton, LinkButton
- Input, Textarea, NumberInput
- Select, Combobox, MultiSelect
- Checkbox, Radio, Switch
- Label, FormField, FormError, FormHint
- Dialog, AlertDialog, Drawer
- DropdownMenu, ContextMenu
- Popover, Tooltip, HoverCard
- Tabs, Accordion, Collapsible
- Toast, Banner, Callout
- Avatar, Badge, Tag, StatusDot
- Progress, Spinner, Skeleton
- Separator, ScrollArea
- Breadcrumb

### Tier 2 — Composed (built on primitives)

Live in `frontend/src/components/common/`.

- `DataTable` — virtualized, sortable, filterable, paginated (the single table pattern)
- `Pagination` — cursor + offset variants
- `FilterBar` — composable filter chips with clearable state
- `SearchInput` — debounced, with keyboard shortcut binding (`/`)
- `EmptyState` — icon + title + description + primary action
- `ErrorState` — error code + message + recovery action
- `LoadingState` — skeleton layout variants
- `FileUploadDropzone` — drag-drop with multipart progress and validation slots
- `LogViewer` — virtualized log stream with search, filter, follow-tail
- `CopyButton` — copy-to-clipboard with confirmation toast
- `DateTime` — renders absolute + relative; timezone aware
- `KeyValueList` — two-column metadata display
- `ConfirmDialog` — wrapper around AlertDialog for destructive actions

### Tier 3 — Domain

Live in `frontend/src/components/domain/`. These encode Senju-specific concepts.

- `JobStatusPill` — queued / running / succeeded / failed / canceled / paused with tokens from `color-palette.md`
- `PipelineStageTimeline` — linear stage visualization with live state
- `StageLogPanel` — per-stage log stream with download
- `VariantRow` — canonical variant presentation (chrom, pos, ref, alt, gene, impact, allele freq)
- `VariantImpactBadge` — ML impact score with confidence band
- `FastqFileCard` — FASTQ metadata + quality summary
- `PipelineRunCard` — job summary card for dashboard
- `UserRoleBadge` — RBAC role visualization
- `ApiTokenRow` — token metadata with reveal/revoke actions

### Tier separation rules

- Primitives never know about domain concepts.
- Composed components may know about primitives and domain-agnostic behavior (tables, forms).
- Domain components may consume both, but must not re-implement primitive behavior.

---

## 10. Icon rules

- **Lucide only.** No other icon sets without an ADR.
- Icon sizes align to spacing: 12 / 14 / 16 / 20 / 24 px.
- Every icon-only button must have `aria-label`. This is lint-enforced.
- Decorative icons must set `aria-hidden="true"`.

---

## 11. Form stack

- State: **React Hook Form**.
- Validation: **Zod**, with schemas generated from or mirroring the OpenAPI spec in `backend/openapi/openapi.yaml`.
- Submission: typed API client from `@/lib/api` (OpenAPI codegen output).
- Errors: field-level errors from server mapped back to fields by name; global errors shown in a `Banner` at the top of the form.

---

## 12. Data-viz rules

- Library TBD in the first viz issue (Recharts or Visx — both MIT).
- All charts read colors from `--color-dv-*` tokens — never hardcoded.
- Every chart has a **data-table fallback** accessible via a "View as table" toggle (a11y requirement).
- Charts must render legible at 320px width (for responsive degradation).

---

## 13. Empty, error, loading — always three states

Every data-driven view must explicitly handle:

1. **Loading** — skeletons matching final layout shape.
2. **Empty** — informative `EmptyState` with a primary action.
3. **Error** — `ErrorState` with retry and (where applicable) support/docs link.

PRs that only implement the happy path are rejected.
