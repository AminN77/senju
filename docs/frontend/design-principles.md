# Design Principles

These are the non-negotiable principles the Senju frontend rolls every decision up to. When two principles conflict, **accessibility wins**, then **familiarity**, then **originality**.

## 1. Familiar, not cloned

Senju's UX should feel immediately usable to someone coming from DNAnexus or a similar commercial genomics platform. We borrow **patterns** (information architecture, navigation shape, workflow mental models, domain terminology) — never **assets** (CSS, component code, icons, illustrations, copywriting, screenshots).

Practical tests:

- "Would a DNAnexus user find the same concepts in the same places?" → Yes.
- "Could a reasonable person, seeing our UI next to DNAnexus's, identify ours as a distinct product?" → Must be Yes.

See `ip-and-references.md` for the full guardrail policy.

## 2. Accessibility is a floor, not a goal

**WCAG 2.1 AA is the minimum** every shipped route must meet. No PR merges without:

- Keyboard reachability for every interactive element.
- Visible focus states on every interactive element (no `outline: none` without a replacement).
- Screen-reader labeling for every icon-only button.
- Contrast ratios meeting AA (4.5:1 body text, 3:1 large text and non-text UI).
- axe-core automated checks passing for the affected routes.

See `accessibility.md`.

## 3. Tokens, not values

No raw hex values, no `px` colors, no hand-typed font sizes in component code. Every visual value is a **token** exposed as a CSS custom property. Themes are token-swaps, not component rewrites.

- If a token doesn't exist for what you need, extend the token system in `design-system.md` and the Tailwind config — don't one-off the value.
- Tailwind arbitrary values for color (`text-[#xxxxxx]`) are **linted out**. Spacing arbitrary values are allowed only with an inline justification comment.

## 4. Dark and light at parity, dark first

Dark mode is the default theme for the authenticated console; light mode must reach full parity from day one. A contrast regression in either theme blocks merge.

## 5. Server-aware, but client-safe

Next.js App Router gives us server components. Data fetching happens on the server where possible. Any component that imports a client-only API (browser globals, stateful hooks) must declare `"use client"` at the top. Server-only secrets must never reach the client bundle — enforced by a lint rule and CI bundle scan.

## 6. Typed contracts, not hopeful strings

The frontend consumes the backend via an OpenAPI-generated TypeScript client. We do not hand-type request/response shapes. Form validation uses Zod schemas derived from (or mirroring) the OpenAPI contract so client-side validation and server-side validation cannot drift silently.

## 7. One pattern per problem

Tables: one table component. Dialogs: one dialog primitive. Forms: one form stack (React Hook Form + Zod). When a contributor is tempted to introduce a second way to do something, that is a signal to improve the first one or write an ADR — not to fork behavior.

## 8. Performance is a feature

Budgets are enforced in CI:

- Marketing route: ≤ 150 KB JS (gzip), LCP ≤ 2.0s on mid-tier laptop cable profile.
- Authenticated console route (dashboard, jobs list): ≤ 250 KB JS (gzip) first load, TTI ≤ 3.0s.
- Variant browser and log viewer: virtualized rendering required (no DOM explosion on 10k+ rows).

Budgets are revisited per ADR, not per PR.

## 9. Copy is UI

Microcopy is part of the design. Sentence case. Active voice. No jargon that a first-time user wouldn't recognize. Error messages say what happened and what to do next. Empty states suggest a first action. Copy lives in a `copy/` module so it can be reviewed independently.

## 10. Boring beats clever

Reach for widely understood libraries and patterns before reaching for novel ones. Optimize for the next contributor's reading speed, not the current contributor's writing speed.
