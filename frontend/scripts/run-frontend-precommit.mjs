import { execSync } from "node:child_process";
import { join } from "node:path";

function repoRoot() {
  return execSync("git rev-parse --show-toplevel", { encoding: "utf8" }).trim();
}

function stagedFiles() {
  const output = execSync("git diff --cached --name-only --diff-filter=ACMR", { encoding: "utf8" }).trim();
  return output ? output.split("\n") : [];
}

const root = repoRoot();
const hasFrontendChanges = stagedFiles().some((file) => file.startsWith("frontend/"));

if (!hasFrontendChanges) {
  console.log("[pre-commit] no frontend staged files; skipping frontend hook checks");
  process.exit(0);
}

const frontendDir = join(root, "frontend");
execSync("pnpm exec lint-staged", { cwd: frontendDir, stdio: "inherit" });
execSync("pnpm typecheck", { cwd: frontendDir, stdio: "inherit" });
execSync("pnpm test:changed", { cwd: frontendDir, stdio: "inherit" });
