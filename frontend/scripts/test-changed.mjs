import { execSync } from "node:child_process";
import { accessSync, constants } from "node:fs";
import { join } from "node:path";

function repoRoot() {
  return execSync("git rev-parse --show-toplevel", { encoding: "utf8" }).trim();
}

function stagedFiles() {
  const output = execSync("git diff --cached --name-only --diff-filter=ACMR", { encoding: "utf8" }).trim();
  return output ? output.split("\n") : [];
}

function fileExists(path) {
  try {
    accessSync(path, constants.F_OK);
    return true;
  } catch {
    return false;
  }
}

const root = repoRoot();
const files = stagedFiles()
  .filter((file) => file.startsWith("frontend/"))
  .filter((file) => /\.(ts|tsx|js|jsx)$/.test(file))
  .map((file) => file.replace(/^frontend\//, ""))
  .filter((file) => fileExists(join(root, "frontend", file)));

if (files.length === 0) {
  console.log("[test:changed] no staged frontend source files; skipping");
  process.exit(0);
}

const quoted = files.map((file) => `"${file}"`).join(" ");
execSync(`pnpm vitest related --run ${quoted}`, {
  cwd: join(root, "frontend"),
  stdio: "inherit",
});
