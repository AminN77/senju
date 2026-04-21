# Component Standards

Rules for authoring, testing, and reviewing frontend components. Every component lives in one of three tiers (primitives, composed, domain) as defined in `design-system.md` §9.

---

## 1. File layout

Each component is a folder, not a single file. This scales as the component grows tests, stories, and variants.

```
frontend/src/components/<tier>/<component-name>/
  index.ts                 # Public re-export only
  <component-name>.tsx     # Component implementation
  <component-name>.types.ts  # Prop types (when non-trivial)
  <component-name>.stories.tsx  # Storybook stories
  <component-name>.test.tsx  # Unit + a11y tests
  <component-name>.module.css # Only if Tailwind is insufficient (rare)
```

Single-file exceptions allowed for trivial wrappers (< 30 LoC, no stateful behavior) — but they still need tests and a story.

---

## 2. Prop API conventions

- Props are **explicit**; no `...rest` pass-through without a declared base type (e.g. `React.ComponentPropsWithoutRef<"button">`).
- Variant props use **string unions**, not booleans that multiply:
  - Good: `variant: "primary" | "secondary" | "ghost" | "danger"`.
  - Bad: `primary?: boolean; secondary?: boolean; ghost?: boolean`.
- Size props: `size: "sm" | "md" | "lg"`. `md` is always the default.
- Event handlers prefixed `on` — `onClick`, `onSelect`, `onValueChange`.
- Controlled / uncontrolled components follow the Radix convention: `value` + `onValueChange` OR `defaultValue`. Never both required.
- `children` typed as `React.ReactNode`. If a component requires a specific child shape, enforce it via a `Compound` subcomponent pattern (e.g. `<Card.Header>`), not via prop drilling.
- Destructive actions are **never** the default. `<ConfirmDialog>` is required for any data-destroying action.

---

## 3. Styling rules

- **Tailwind utilities** for layout, spacing, typography, color.
- **No raw hex values** in `className` or inline styles. Lint-enforced.
- **No inline `style={{}}`** for design-system values. `style` is allowed only for dynamic runtime values (e.g. a progress bar width).
- Variants are composed via `class-variance-authority` (`cva`) — lives in `<component>.tsx`, not inline conditionals.
- Theme-aware classes reference tokens: `bg-surface-raised text-text-primary border-border-default`.

---

## 4. Server vs client components

- Default to **server components** (no `"use client"`).
- Add `"use client"` **only** when the component uses: stateful hooks, browser-only APIs, event handlers, Context, or third-party client-only libs.
- Never import from `server-only` code inside a client component (enforced by `server-only` / `client-only` package markers).
- Do not reach for `"use client"` on the whole page to avoid a thinking — push client boundaries as deep as possible.

---

## 5. Accessibility

See `accessibility.md` for the full bar. For components specifically:

- Every interactive element reachable via keyboard.
- `aria-label` on icon-only buttons — **lint-enforced**.
- Focus state visible by default; never remove without replacement.
- `jest-axe` assertion in the `.test.tsx` file for every interactive component:

```tsx
it("has no a11y violations", async () => {
  const { container } = render(<Component />);
  expect(await axe(container)).toHaveNoViolations();
});
```

---

## 6. Testing

| Type | Tool | Required for |
| --- | --- | --- |
| Unit / interaction | Vitest + React Testing Library | Every component |
| A11y | jest-axe (as above) | Every interactive component |
| Visual / states | Storybook stories | Every component |
| Route / E2E | Playwright | Every shipped route |
| Contrast | `pnpm test:contrast` | Token changes |

**Coverage is not a target**; meaningful behavior coverage is. Trivial render-smoke tests are not enough for a component with real branching.

Tests live next to the component. Integration tests for multi-component flows live under `frontend/tests/integration/`.

---

## 7. Storybook

Every component has stories covering:

- **Default** state.
- Each **variant** × **size**.
- **Loading**, **empty**, **error** states where applicable.
- **Disabled** / **read-only** if supported.
- **Long content** overflow (truncation behavior).
- **RTL** (document that RTL is out-of-phase but the story should not break).

Stories double as the a11y test surface (`@storybook/addon-a11y`) and the visual-regression surface (Chromatic or Loki — decision deferred to the Storybook setup issue).

---

## 8. Naming

- Component folder and file: `kebab-case` (`job-status-pill/`).
- Component export: `PascalCase` (`JobStatusPill`).
- Types: `PascalCase`, prop types suffixed `Props` (`JobStatusPillProps`).
- Hooks: `camelCase`, prefixed `use` (`useJobPolling`).
- Utility modules: `camelCase`.

---

## 9. Performance

- **Memoize by default only when measured.** `React.memo` / `useMemo` / `useCallback` solve specific re-render problems — profile first.
- Virtualize any list > 100 rows (use `@tanstack/react-virtual`).
- Lazy-load routes not on the critical path via `next/dynamic` with loading skeletons.
- Avoid importing whole icon sets — `lucide-react` is tree-shakeable; import per icon.

---

## 10. Documentation

Every component exports a JSDoc block describing:

- What the component is for (one sentence).
- When **not** to use it (pointer to the right alternative).
- Any a11y contract the consumer must uphold (e.g. "requires a parent `<DialogTrigger>`").

Component docs are auto-aggregated into Storybook's docs tab; no separate `.md` files per component.

---

## 11. Review checklist

Reviewers of a new or changed component verify:

- [ ] Props API follows conventions (§2).
- [ ] Uses tokens only — no raw hex, no arbitrary spacing values without justification (§3).
- [ ] Server/client boundary is correct (§4).
- [ ] A11y: keyboard path, focus visibility, screen-reader labeling, jest-axe passes (§5).
- [ ] Tests cover real branches, not just render smoke (§6).
- [ ] Storybook covers all variants + all states + long content (§7).
- [ ] No assets copied from third parties (see `ip-and-references.md`).
- [ ] Performance: virtualized if list > 100, lazy-loaded if off critical path (§9).
- [ ] JSDoc present (§10).
