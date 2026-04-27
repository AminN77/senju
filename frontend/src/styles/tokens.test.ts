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
    expect(
      resolveValue(lightStyles.getPropertyValue("--color-surface-raised")).replace(/\s+/g, "")
    ).toBe("oklch(100%00)");
    expect(
      resolveValue(lightStyles.getPropertyValue("--color-text-primary")).replace(/\s+/g, "")
    ).toBe("oklch(20%0.014250)");

    setTheme("dark");
    const darkStyles = getComputedStyle(document.documentElement);
    expect(
      resolveValue(darkStyles.getPropertyValue("--color-surface-raised")).replace(/\s+/g, "")
    ).toBe("oklch(20%0.014250)");
    expect(
      resolveValue(darkStyles.getPropertyValue("--color-text-primary")).replace(/\s+/g, "")
    ).toBe("oklch(98%0.003250)");
  });

  it("defines typography scale tokens from the design spec", () => {
    const expectedTypographyTokens: Array<[token: string, value: string]> = [
      ["text-display-lg", "3rem"],
      ["text-display-md", "2.25rem"],
      ["text-heading-xl", "1.75rem"],
      ["text-heading-lg", "1.375rem"],
      ["text-heading-md", "1.125rem"],
      ["text-heading-sm", "1rem"],
      ["text-body-lg", "1rem"],
      ["text-body-md", "0.875rem"],
      ["text-body-sm", "0.8125rem"],
      ["text-caption", "0.75rem"],
      ["text-code-md", "0.8125rem"],
      ["text-code-sm", "0.75rem"],
      ["leading-display-lg", "3.5rem"],
      ["leading-display-md", "2.75rem"],
      ["leading-heading-xl", "2.25rem"],
      ["leading-heading-lg", "1.875rem"],
      ["leading-heading-md", "1.625rem"],
      ["leading-heading-sm", "1.5rem"],
      ["leading-body-lg", "1.625rem"],
      ["leading-body", "1.375rem"],
      ["leading-body-sm", "1.25rem"],
      ["leading-caption", "1.125rem"],
      ["leading-code-md", "1.25rem"],
      ["leading-code-sm", "1.125rem"],
      ["tracking-display", "-0.02em"],
      ["tracking-heading", "-0.01em"],
      ["tracking-body", "0em"],
      ["tracking-caption", "0.01em"],
      ["tracking-code", "0em"],
    ];

    for (const [token, expected] of expectedTypographyTokens) {
      const match = tokensCss.match(new RegExp(`--${token}:\\s*([^;]+);`));
      expect(match?.[1]?.trim()).toBe(expected);
    }
  });
});
