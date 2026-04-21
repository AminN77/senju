# ADR-0005: Frontend framework (Next.js + TypeScript)

- **Status:** Accepted
- **Date:** 2026-04-21
- **Owners:** Platform team
- **Decision scope:** The user-facing web application for Senju, covering both the public marketing surface and the authenticated console (dashboard, pipeline jobs, variant browser, ML scoring, admin, settings).

## Context

The backend is established (Go + Gin + Postgres + ClickHouse + MinIO + NATS, per ADR-0001–0004). There is no frontend yet. The first UI phase must deliver:

1. A **public marketing page** for promotion and OSS adoption (needs SEO-friendly rendering).
2. An **authenticated console** for uploads, pipeline jobs, variant browsing, ML impact scoring, admin, and settings.
3. **One codebase, one deployment** to keep the project operationally small.

The backend exposes **REST + OpenAPI 3** (ADR-0004), so the frontend must generate a typed client from `backend/openapi/openapi.yaml`.

The project's primary UX reference is **DNAnexus** (for information architecture and workflow patterns only — visual identity is original; see `docs/frontend/ip-and-references.md`).

## Decision

1. **Framework:** Use **[Next.js](https://nextjs.org/) (latest stable, App Router)** with **React** and **TypeScript (strict mode)**.
2. **Rendering:** **SSR/SSG** for marketing and public routes; **Client-rendered + streaming** for authenticated console routes. Route segments declare their rendering mode explicitly.
3. **Routing:** App Router (`app/` directory) with route groups `(marketing)` and `(app)` to split public and authenticated surfaces.
4. **Language:** TypeScript **strict mode** is non-negotiable. No `any` without an inline justification comment; no implicit `any`.
5. **Package manager:** **pnpm** (fast, disk-efficient, good monorepo ergonomics should we split `packages/` later).
6. **Node runtime:** Pin Node LTS via `.nvmrc` / `engines` field.

## Alternatives considered

1. **Vite + React SPA**
   - Pros: simpler mental model, faster dev server for pure SPAs.
   - Cons: no first-class SSR/SSG for the marketing page; would force a second static-site tool or a hybrid that splits our codebase.
2. **Remix**
   - Pros: excellent data-loading primitives, standards-aligned.
   - Cons: smaller ecosystem; shadcn/ui and most community templates center on Next.js.
3. **SvelteKit**
   - Pros: smaller bundles, elegant.
   - Cons: smaller talent pool and component ecosystem; higher onboarding cost for contributors.
4. **Two apps (static marketing + SPA console)**
   - Pros: separation of concerns.
   - Cons: double the CI, auth-sharing complexity, double the contributor surface for no benefit at MVP.

## Trade-offs

- **Gain:** One framework covers marketing + console; strong ecosystem; native App Router supports code-splitting, streaming, and route-level rendering modes; well-understood by contributors.
- **Cost:** Next.js opinions (file-system routing, React Server Components) require team discipline to avoid leaking server-only code into the client bundle.
- **Constraint:** Contributors must understand App Router server/client boundaries; we will enforce this via lint rules and code review.

## Risks

- **Bundle bloat** if server/client boundaries are mishandled — mitigated by `"use client"` discipline and bundle-size budgets in CI.
- **Vendor drift** (Vercel-specific APIs) — mitigated by avoiding Vercel-only features; deploy target is a container.

## Migration path

- If we later split marketing to a static site, the App Router's `(marketing)` route group can be extracted to a separate project without touching the console.
- If we abandon Next.js, the React components, Tailwind styles, design tokens, and OpenAPI-generated client are portable.

## Consequences

The frontend ships as a single Next.js + TypeScript app under `frontend/`, with SSR for public pages and a client-rendered authenticated console. Deployment is containerized (consistent with the existing `docker-compose.yml` pattern). Contributors target a single, well-documented stack.
