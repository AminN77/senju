# Senju Frontend

Next.js (App Router) + React + TypeScript (strict) scaffold for the Senju UI Foundation phase.

## Requirements

- Node.js 22 LTS (`.nvmrc`)
- pnpm 10+

## Scripts

- `pnpm dev` - starts local development server on port 3000
- `pnpm build` - production build
- `pnpm start` - starts production server
- `pnpm lint` - ESLint checks with zero warnings allowed
- `pnpm typecheck` - strict TypeScript checks
- `pnpm format` - Prettier formatting check

## Quick start

```bash
pnpm install
pnpm dev
```

Open <http://localhost:3000>.

## Notes

- Next.js config uses `output: "standalone"` for containerized deploys.
- API contract types will be generated from `backend/openapi/openapi.yaml` in later issues.
