# Frontend Dependency Licenses

This file tracks **direct** runtime and development dependencies used by the `frontend/` app and their **SPDX**-style license identifiers, as read from each package’s `package.json` at the resolved install path.

## Policy

- Every dependency change must update this file.
- Only license-compatible dependencies are allowed for OSS distribution.
- If attribution is required by a license, include it here.

## Direct dependencies (direct `dependencies` and `devDependencies` in `frontend/package.json`)

| Package | Version | License (SPDX) | Kind |
| --- | --- | --- | --- |
| `@radix-ui/react-label` | 2.1.8 | MIT | runtime |
| `@radix-ui/react-primitive` | 2.1.4 | MIT | runtime |
| `@radix-ui/react-separator` | 1.1.8 | MIT | runtime |
| `@radix-ui/react-slot` | 1.2.4 | MIT | runtime |
| `@types/node` | 22.19.17 | MIT | dev |
| `@types/react` | 19.2.14 | MIT | dev |
| `@types/react-dom` | 19.2.3 | MIT | dev |
| `autoprefixer` | 10.5.0 | MIT | dev |
| `class-variance-authority` | 0.7.1 | Apache-2.0 | runtime |
| `clsx` | 2.1.1 | MIT | runtime |
| `culori` | 4.0.2 | MIT | dev |
| `eslint` | 9.39.4 | MIT | dev |
| `eslint-config-next` | 16.2.4 | MIT | dev |
| `eslint-config-prettier` | 10.1.8 | MIT | dev |
| `eslint-plugin-jsx-a11y` | 6.10.2 | MIT | dev |
| `jsdom` | 29.0.2 | MIT | dev |
| `lucide-react` | 1.8.0 | ISC | runtime |
| `next` | 16.2.4 | MIT | runtime |
| `postcss` | 8.5.10 | MIT | dev |
| `prettier` | 3.8.3 | MIT | dev |
| `react` | 19.2.4 | MIT | runtime |
| `react-dom` | 19.2.4 | MIT | runtime |
| `shadcn` | 4.4.0 | MIT | dev |
| `tailwind-merge` | 3.5.0 | MIT | runtime |
| `tailwindcss` | 3.4.17 | MIT | dev |
| `typescript` | 5.9.3 | Apache-2.0 | dev |
| `vitest` | 4.1.5 | MIT | dev |

## Maintenance

- After adding, removing, or bumping a direct dependency, refresh this table. You can use `pnpm list --depth 0 --json` to list resolved install paths, then read each package’s `package.json` for `version` and `license`.
