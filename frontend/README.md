# Senju Frontend

Next.js (App Router) + React + TypeScript (strict) scaffold for the Senju UI Foundation phase.

## Requirements

- Node.js 22 LTS (`.nvmrc`)
- pnpm 10+
- Docker + Docker Compose (preferred execution path)

## Scripts

- `pnpm dev` - starts local development server on port 3000
- `pnpm build` - production build
- `pnpm start` - starts production server
- `pnpm lint` - ESLint checks with zero warnings allowed
- `pnpm typecheck` - strict TypeScript checks
- `pnpm format` - Prettier formatting check

## Quick start (Docker first)

```bash
docker compose up -d --build frontend api
```

Open <http://localhost:3001> (or `FRONTEND_PORT` if overridden).

## Verification via Docker

```bash
docker compose run --rm frontend pnpm lint
docker compose run --rm frontend pnpm typecheck
docker compose run --rm frontend pnpm build
```

## Optional local workflow

If you need direct local iteration instead of containers:

```bash
pnpm install
pnpm dev
```

Open <http://localhost:3000>.

## Notes

- Next.js config uses `output: "standalone"` for containerized deploys.
- API contract types will be generated from `backend/openapi/openapi.yaml` in later issues.
