# Frontend Contribution Notes

## Adding shadcn/ui primitives

This project keeps shadcn component source in-repo under `src/components/ui/`.

1. Confirm you are in `frontend/`.
2. Run:

```bash
pnpm shadcn add <component>
```

Example:

```bash
pnpm shadcn add textarea
```

## Expectations after generating a component

- **Stories are required**: Every new component must include Storybook stories for default, variants × sizes (when applicable), loading/empty/error, disabled, and long-content states before review.
- **Story checklist**: Confirm each story file covers `default`, `variants×sizes`, `loading`, `empty`, `error`, `disabled`, and `long content` (mark `not applicable` explicitly when a primitive does not support a state).
- **Source layout**: Generated files live in `src/components/ui/` (or the path the generator outputs to).
- **Tokens**: Use token-backed utilities only (no raw color literals, no Tailwind arbitrary color values for design system colors).
- **Storybook**: Add or update a story for the default state and the main visual / interaction variants you rely on in product.
- **Accessibility**: Add an automated axe-based accessibility test (e.g. jest-axe, or the Vitest-appropriate equivalent if the repo uses Vitest) for the new or changed component surface, and add a short **JSDoc** on the public component that describes the accessibility contract (labels, focus order, roles, live regions) as it applies to that component.
- **Design system review**: Get approval from a design-system owner before merge when the change adds or meaningfully alters a shared UI primitive.
- **Verification** (run locally before opening or updating a PR):

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm test:contrast
```

## Pre-commit hooks

This repo uses a root git hook (`.githooks/pre-commit`) to provide fast local feedback.

### One-time setup

From `frontend/`, run:

```bash
pnpm hooks:install
```

This configures `core.hooksPath` to the committed `.githooks/` directory.

### What runs on commit

- If staged files are only outside `frontend/**`, frontend hooks are skipped.
- If staged files include `frontend/**`, the hook runs:
  - `lint-staged` (ESLint + Prettier on staged files)
  - `pnpm typecheck`
  - `pnpm test:changed` (Vitest related mode for staged source changes)

### Bypass

`git commit --no-verify` bypasses hooks. Use it only for emergency/unblocking workflows (for example, temporarily committing WIP to recover from a local environment failure). Do **not** use it to bypass legitimate lint/type/test failures.
