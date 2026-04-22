import { readFileSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

describe("theme tokens", () => {
  const tokensPath = join(process.cwd(), "src/styles/tokens.css");
  const tokensCss = readFileSync(tokensPath, "utf8");

  const setTheme = (theme: "light" | "dark"): void => {
    document.documentElement.setAttribute("data-theme", theme);
  };

  const resolveValue = (value: string): string => {
    return value.replace(/var\(--([^)]+)\)/g, (_, nestedName: string) => {
      const rootMatch = tokensCss.match(new RegExp(`--${nestedName}:\\s*([^;]+);`));
      if (!rootMatch?.[1]) {
        return "";
      }
      return rootMatch[1].trim();
    });
  };

  it("resolves expected CSS variables for light and dark themes", () => {
    const style = document.createElement("style");
    style.textContent = tokensCss;
    document.head.appendChild(style);

    setTheme("light");
    const lightStyles = getComputedStyle(document.documentElement);
    expect(resolveValue(lightStyles.getPropertyValue("--color-surface-raised")).replace(/\s+/g, "")).toBe(
      "oklch(100%00)"
    );
    expect(resolveValue(lightStyles.getPropertyValue("--color-text-primary")).replace(/\s+/g, "")).toBe(
      "oklch(20%0.014250)"
    );

    setTheme("dark");
    const darkStyles = getComputedStyle(document.documentElement);
    expect(resolveValue(darkStyles.getPropertyValue("--color-surface-raised")).replace(/\s+/g, "")).toBe(
      "oklch(20%0.014250)"
    );
    expect(resolveValue(darkStyles.getPropertyValue("--color-text-primary")).replace(/\s+/g, "")).toBe(
      "oklch(98%0.003250)"
    );
  });
});
