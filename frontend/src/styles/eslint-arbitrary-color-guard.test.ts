import { execSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";

describe("eslint arbitrary color guard", () => {
  const tempDirs: string[] = [];

  afterEach(() => {
    while (tempDirs.length > 0) {
      const dir = tempDirs.pop();
      if (dir) {
        rmSync(dir, { recursive: true, force: true });
      }
    }
  });

  it("blocks Tailwind arbitrary color utilities", () => {
    const dir = mkdtempSync(join(process.cwd(), "src/styles/.tmp-eslint-"));
    tempDirs.push(dir);

    const fixturePath = join(dir, "fixture.tsx");
    const arbitraryClass = "text-" + "[#ff0000]";
    writeFileSync(
      fixturePath,
      `export function Demo(){return <div className="${arbitraryClass}" />;}`
    );

    let output = "";
    let exitCode = 0;
    try {
      execSync(`pnpm eslint "${fixturePath}"`, {
        cwd: process.cwd(),
        encoding: "utf8",
        stdio: "pipe",
      });
    } catch (error) {
      const commandError = error as { stdout?: string; stderr?: string; status?: number };
      output = `${commandError.stdout ?? ""}\n${commandError.stderr ?? ""}`;
      exitCode = commandError.status ?? 1;
    }

    expect(exitCode).toBe(1);
    expect(output.includes("Do not use Tailwind arbitrary color values")).toBe(true);
  });
});
