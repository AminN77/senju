import { readFileSync } from "node:fs";
import { join } from "node:path";

import { converter, wcagContrast } from "culori";
import { describe, expect, it } from "vitest";

const toRgb = converter("rgb");
const tokensCss = readFileSync(join(process.cwd(), "src/styles/tokens.css"), "utf8");

const blockForTheme = (theme: "light" | "dark"): string => {
  if (theme === "light") {
    const [, lightBlock = ""] =
      tokensCss.match(/:root\s*\{([\s\S]*?)\}\s*\[data-theme="dark"\]/) ?? [];
    return lightBlock;
  }
  const [, darkBlock = ""] = tokensCss.match(/\[data-theme="dark"\]\s*\{([\s\S]*?)\}/) ?? [];
  return darkBlock;
};

const extractToken = (block: string, tokenName: string): string => {
  const tokenPattern = new RegExp(`--${tokenName}:\\s*([^;]+);`);
  const match = block.match(tokenPattern);
  const tokenValue = match?.[1];
  if (!tokenValue) {
    throw new Error(`Token --${tokenName} not found`);
  }
  return tokenValue.trim();
};

const parseColor = (value: string, theme: "light" | "dark"): string => {
  const themeBlock = blockForTheme(theme);
  const lightBlock = blockForTheme("light");
  const resolved = value.replace(/var\(--([^)]+)\)/g, (_, nestedName: string) => {
    try {
      return extractToken(themeBlock, nestedName);
    } catch {
      return extractToken(lightBlock, nestedName);
    }
  });

  const rgb = toRgb(resolved);
  if (!rgb) {
    throw new Error(`Unable to parse color value: ${value}`);
  }
  return resolved;
};

describe("contrast checks", () => {
  const tokenValueForTheme = (theme: "light" | "dark", tokenName: string): string => {
    const themeBlock = blockForTheme(theme);
    const lightBlock = blockForTheme("light");
    try {
      return extractToken(themeBlock, tokenName);
    } catch {
      return extractToken(lightBlock, tokenName);
    }
  };

  const bodyTextPairs = [
    ["color-text-primary", "color-surface-base"],
    ["color-text-primary", "color-surface-raised"],
    ["color-text-secondary", "color-surface-base"],
    ["color-text-secondary", "color-surface-raised"],
    ["color-text-on-inverse", "color-surface-inverse"],
  ] as const;
  const nonTextPairs = [
    ["color-border-focus", "color-surface-base"],
    ["color-success-solid", "color-surface-base"],
    ["color-warning-solid", "color-surface-base"],
    ["color-danger-solid", "color-surface-base"],
    ["color-info-solid", "color-surface-base"],
  ] as const;

  for (const theme of ["light", "dark"] as const) {
    it(`ensures ${theme} theme body-text pairings meet 4.5:1 contrast`, () => {
      const block = blockForTheme(theme);

      for (const [foregroundName, backgroundName] of bodyTextPairs) {
        const foreground = parseColor(extractToken(block, foregroundName), theme);
        const background = parseColor(tokenValueForTheme(theme, backgroundName), theme);
        const contrast = wcagContrast(foreground, background);
        expect(
          contrast,
          `Expected --${foregroundName} contrast ${contrast.toFixed(2)} to be >= 4.50 against --${backgroundName} (${theme})`
        ).toBeGreaterThanOrEqual(4.5);
      }
    });

    it(`ensures ${theme} theme non-text pairings meet 3:1 contrast`, () => {
      const block = blockForTheme(theme);

      for (const [foregroundName, backgroundName] of nonTextPairs) {
        const foreground = parseColor(extractToken(block, foregroundName), theme);
        const background = parseColor(extractToken(block, backgroundName), theme);
        const contrast = wcagContrast(foreground, background);
        expect(
          contrast,
          `Expected --${foregroundName} contrast ${contrast.toFixed(2)} to be >= 3.00 against --${backgroundName} (${theme})`
        ).toBeGreaterThanOrEqual(3);
      }
    });

    it(`ensures ${theme} theme data-viz categorical tokens meet 3:1 contrast`, () => {
      const block = blockForTheme(theme);
      const surfaceBase = parseColor(extractToken(block, "color-surface-base"), theme);

      for (let i = 1; i <= 10; i += 1) {
        const tokenValue = parseColor(extractToken(block, `color-dv-${i}`), theme);
        const contrast = wcagContrast(tokenValue, surfaceBase);
        expect(
          contrast,
          `Expected --color-dv-${i} contrast ${contrast.toFixed(2)} to be >= 3.00 against ${theme} surface`
        ).toBeGreaterThanOrEqual(3);
      }
    });
  }
});
