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
