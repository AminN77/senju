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

- Generated files live in `src/components/ui/`.
- Components use token-backed utilities (no raw color literals).
- Run verification:

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm test:contrast
```
