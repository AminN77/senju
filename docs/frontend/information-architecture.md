# Information Architecture

Senju's IA borrows the **mental model** that users of commercial genomics platforms (DNAnexus in particular) already carry. The goal is familiarity-of-concept, not visual duplication (see `ip-and-references.md`).

---

## 1. Route map

Top-level route groups in the Next.js App Router:

```
app/
  (marketing)/
    page.tsx                         GET /
    pricing/page.tsx                 GET /pricing          [deferred — out of phase]
    docs/page.tsx                    GET /docs             [link out to docs site]
    changelog/page.tsx               GET /changelog        [deferred — out of phase]
  (auth)/
    login/page.tsx                   GET /login
    signup/page.tsx                  GET /signup
    forgot-password/page.tsx         GET /forgot-password
    reset-password/page.tsx          GET /reset-password
    verify-email/page.tsx            GET /verify-email
  (app)/
    dashboard/page.tsx               GET /dashboard        [default authenticated landing]
    projects/
      page.tsx                       GET /projects
      [projectId]/
        page.tsx                     GET /projects/:id
        files/page.tsx               GET /projects/:id/files
        upload/page.tsx              GET /projects/:id/upload
    jobs/
      page.tsx                       GET /jobs
      [jobId]/
        page.tsx                     GET /jobs/:id         [overview]
        stages/[stageId]/page.tsx    GET /jobs/:id/stages/:stageId
        logs/page.tsx                GET /jobs/:id/logs
        artifacts/page.tsx           GET /jobs/:id/artifacts
    variants/
      page.tsx                       GET /variants          [query UI]
      [variantId]/page.tsx           GET /variants/:id      [detail drawer or page]
    ml/
      impact/page.tsx                GET /ml/impact         [scoring UI]
      impact/models/page.tsx         GET /ml/impact/models
    admin/
      users/page.tsx                 GET /admin/users
      roles/page.tsx                 GET /admin/roles
      tokens/page.tsx                GET /admin/tokens
      audit/page.tsx                 GET /admin/audit
    settings/
      profile/page.tsx               GET /settings/profile
      security/page.tsx              GET /settings/security
      api-keys/page.tsx              GET /settings/api-keys
      preferences/page.tsx           GET /settings/preferences
```

---

## 2. Navigation shape

Two surfaces — marketing and authenticated — have distinct shells.

### Marketing shell

- **Top bar**: product name, docs link, pricing link (hidden until enabled), GitHub link, "Log in" (secondary), "Get started" (primary CTA).
- **Footer**: product / resources / company / legal, OSS license badge, social links.

### Authenticated app shell

- **Primary sidebar** (collapsible): Dashboard · Projects · Jobs · Variants · ML Impact · Admin · Settings · (workspace switcher at the top).
- **Top bar**: breadcrumb, global search (`/` shortcut), notifications (deferred), user menu (profile · sign out · theme toggle).
- **Content region**: page title · optional tabs · primary actions (right-aligned) · content.
- **Right-side panel** (contextual): log streams, filter drawers, variant detail — invoked as needed, not persistent.

### Keyboard shortcuts (global)

| Shortcut | Action |
| --- | --- |
| `/` | Focus global search |
| `g d` | Go to Dashboard |
| `g j` | Go to Jobs |
| `g v` | Go to Variants |
| `g p` | Go to Projects |
| `?` | Open shortcut cheat sheet |
| `Esc` | Close dialogs, drawers, menus |

Shortcuts documented in a `?` dialog; never activate when focus is in an input.

---

## 3. Page inventory (UI Foundation phase)

Each entry is a **promise of what ships** — acceptance criteria in the corresponding GitHub issue elaborate further.

### Marketing `/`

- Hero: product statement, primary CTA (Get started), secondary (View on GitHub).
- Value props (3–4 sections with illustrations — original, not sourced from competitors).
- Architecture glance (high-level diagram of data flow — original diagram).
- "Built on open source" section with the stack.
- Footer with license, repo link, docs link.

### Auth `/login`, `/signup`, `/forgot-password`, `/reset-password`, `/verify-email`

- Email + password.
- SSO placeholder (OIDC button — behind feature flag, deferred to integration phase).
- Error states: locked account, unverified email, expired reset token.
- Rate-limit feedback surfaced from backend.

### Dashboard `/dashboard`

- Greeting + current workspace context.
- Summary stat cards: active jobs, queued jobs, failed jobs in last 24h, storage used.
- Recent jobs list (top 10, links to `/jobs/:id`).
- Recent uploads (top 10, links to file detail).
- Quick actions: upload FASTQ, run pipeline, browse variants.

### Projects `/projects`, `/projects/:id`, `/projects/:id/files`, `/projects/:id/upload`

- Projects list: name, owner, created, job count.
- Project detail: metadata + tabs (Files · Jobs · Settings).
- Files: virtualized `DataTable` with type, size, checksum, status.
- Upload: `FileUploadDropzone`, multipart progress, validation feedback, presign flow.

### Jobs `/jobs`, `/jobs/:id`, `/jobs/:id/stages/:stageId`, `/jobs/:id/logs`, `/jobs/:id/artifacts`

- Jobs list: status, pipeline, project, started, duration, owner. Filters on status, pipeline, project, date range.
- Job overview: metadata, `PipelineStageTimeline`, summary metrics, quick links to logs and artifacts.
- Stage detail: inputs, outputs, resource usage, exit code, stage-level logs.
- Logs: `LogViewer` with search, level filter, follow-tail, download.
- Artifacts: file listing with download, checksum, preview where safe.

### Variants `/variants`, `/variants/:id`

- Query UI: filter bar (chromosome, position range, gene, consequence, allele frequency, quality, impact score).
- Results: virtualized `DataTable` with `VariantRow` presentation.
- Variant detail: annotations, allele context, linked jobs, exportable row.

### ML Impact `/ml/impact`, `/ml/impact/models`

- Scoring UI: input variants (upload or paste), model selection, run scoring.
- Results: impact score with `VariantImpactBadge`, confidence interval, contributing features.
- Models: model registry, version, training metadata, performance metrics.

### Admin `/admin/users`, `/admin/roles`, `/admin/tokens`, `/admin/audit`

- Users: list with role, last active, status; invite flow; deactivate.
- Roles: role matrix (RBAC from backend security baseline); read-only in phase 1 except where backend supports role edits.
- API tokens: create, reveal-once, revoke, list with metadata.
- Audit log: read-only event stream with filters.

### Settings `/settings/profile`, `/settings/security`, `/settings/api-keys`, `/settings/preferences`

- Profile: name, email (verify flow), avatar (optional, deferred).
- Security: password change, MFA status (toggle gated on backend support), session list with revoke.
- API keys: personal tokens (mirrors `/admin/tokens` shape but self-scoped).
- Preferences: theme (auto / dark / light), timezone, date format, keyboard shortcuts toggle.

---

## 4. URL and state conventions

- **Query params are canonical for filter state.** `/jobs?status=running&pipeline=gatk&from=2026-04-01` — always shareable, always reloadable.
- **Path params for entity identity.** `/jobs/:id` — never pass the id via query.
- **Cursor-based pagination** preferred for large collections; surface the cursor as `?after=...`. Offset pagination allowed for admin/audit where total count matters.
- **Back button must always work.** Dialogs/drawers that hold important state push URL state; trivial menus do not.

---

## 5. Empty-repo, empty-workspace, empty-project states

These are the ones users see on day one. Each must actively guide the next action:

- **Empty dashboard:** "Create your first project" CTA + "Upload FASTQ" CTA + link to docs.
- **Empty project (no files):** "Upload FASTQ" with the dropzone inline.
- **Empty jobs list:** "No jobs yet — start a pipeline" with pipeline picker.
- **Empty variants:** "No variants — run a pipeline or adjust your filters."

No marketing-style hero in empty states inside the app. Empty states are functional, not aspirational.
