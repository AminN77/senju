import { execSync } from "node:child_process";
import { join } from "node:path";

const root = execSync("git rev-parse --show-toplevel", { encoding: "utf8" }).trim();
const hooksPath = join(root, ".githooks");

execSync(`git config core.hooksPath "${hooksPath}"`, { stdio: "inherit" });
console.log(`[hooks] git hooks path configured: ${hooksPath}`);
