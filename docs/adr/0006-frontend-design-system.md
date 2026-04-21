# ADR-0006: Frontend design system (shadcn/ui + Radix + Tailwind + CSS-variable tokens)

- **Status:** Accepted
- **Date:** 2026-04-21
- **Owners:** Platform team
- **Decision scope:** Component foundation, styling strategy, theming, and token system for the Senju frontend.

## Context

ADR-0005 selects Next.js + TypeScript. We need:

- An **accessible** component foundation (WCAG 2.1 AA is the project minimum; see `docs/frontend/accessibility.md`).
- **Full ownership of component source** so we can evolve visual identity without fighting a vendor's opinions.
- **Dark + light themes at parity** from day one, driven by tokens (not per-component overrides).
- A license posture that is safe for an **MIT/open-source** distribution.
- A foundation that does **not** require copying any competitor's CSS, components, icons, or proprietary assets (see `docs/frontend/ip-and-references.md`).

## Decision

1. **Primitives:** **[Radix UI](https://www.radix-ui.com/primitives)** for unstyled, accessible component behavior (dialog, dropdown, tabs, popover, tooltip, select, etc.).
2. **Component layer:** **[shadcn/ui](https://ui.shadcn.com/)** — components are generated into our repo (`frontend/src/components/ui/`) as source we own, not a package dependency. MIT-licensed.
3. **Styling:** **[Tailwind CSS](https://tailwindcss.com/)** (latest stable) with a **token-first** configuration. No arbitrary hex values in component code — every color, spacing, radius, and shadow reads from a token.
4. **Theme mechanism:** **CSS custom properties** (CSS variables) on `:root` and `[data-theme="dark"]`. Tailwind's theme reads from these variables so the same class compiles to theme-aware styles.
5. **Icons:** **[Lucide](https://lucide.dev/)** (MIT, tree-shakeable). No proprietary icon sets.
6. **Fonts:** **Inter** (UI) and **JetBrains Mono** (code/logs), both OFL-licensed, loaded via `next/font` with subsetting.
7. **Data viz:** **[Recharts](https://recharts.org/)** or **[Visx](https://airbnb.io/visx/)** — decision deferred to the first data-viz issue; both are MIT.
8. **Forms:** **React Hook Form** + **Zod** for schema validation (types derived from OpenAPI where possible).
9. **Component testing:** Components must have **accessibility tests** (`@testing-library/react` + `jest-axe`) and **visual stories** (Storybook).

## Alternatives considered

1. **Material UI (MUI)**
   - Pros: mature, batteries-included.
   - Cons: strong opinionated visual identity, hard to theme away from Material; larger bundle; MUI Pro components are not MIT.
2. **Chakra UI v3**
   - Pros: good DX, accessible.
   - Cons: theming system is less portable than Tailwind+tokens; fewer domain-table patterns for the data surface we need.
3. **Mantine**
   - Pros: rich component library, accessible.
   - Cons: coupling to Mantine's theme object; harder to evolve visual identity without deep engagement with Mantine internals.
4. **Radix + Tailwind, hand-rolled (no shadcn)**
   - Pros: no generated code.
   - Cons: re-implements what shadcn already composes correctly; slower to start.
5. **Build everything from scratch**
   - Pros: complete ownership.
   - Cons: accessibility regressions are near-certain; a11y for complex primitives (combobox, dialog focus-traps, roving tabindex) is a solved problem we shouldn't re-solve.

## Trade-offs

- **Gain:** We own component source (shadcn copies files in, not a package), Radix gives us a11y for free, Tailwind + CSS variables give us cheap dark/light parity, everything is MIT.
- **Cost:** shadcn components live in our repo, so updates are opt-in (we must track upstream changes manually). This is the intended trade — ownership over auto-updates.
- **Constraint:** Components in `frontend/src/components/ui/` MUST be modifiable by us; we do not treat them as vendor code.

## Risks

- **Token drift** if contributors bypass tokens (e.g., write `text-[#ff0000]`) — mitigated by an ESLint rule (`no-restricted-syntax` on Tailwind arbitrary values for colors) and `docs/frontend/design-system.md` enforcement.
- **Dark-mode contrast bugs** — mitigated by automated contrast checks in CI (see `docs/frontend/accessibility.md`).
- **Accidental IP contamination** from copying competitor CSS — mitigated by `.cursor/rules/frontend-ip-guardrails.mdc` and review policy in `docs/frontend/ip-and-references.md`.

## Migration path

- If we drop shadcn, we retain the Radix + Tailwind + token foundation; only the generated component files change.
- If we drop Tailwind, tokens remain valid because they live as CSS variables; we'd rewrite class usage only.

## Consequences

The design system is token-driven, theme-aware, accessible by default, and entirely MIT-licensed. Component source lives in our repo. Contributors style via tokens and Tailwind utilities, never with raw hex values or inline styles for colors/spacing/typography.
