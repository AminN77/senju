import { defineConfig, globalIgnores } from "eslint/config";
import jsxA11y from "eslint-plugin-jsx-a11y";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";
import eslintConfigPrettier from "eslint-config-prettier";

const jsxA11yErrorRules = Object.fromEntries(
  Object.keys(jsxA11y.flatConfigs.recommended.rules).map((ruleName) => [ruleName, "error"])
);

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    files: ["**/*.{js,jsx,mjs,cjs,ts,tsx}"],
    rules: {
      ...jsxA11yErrorRules,
      "no-restricted-syntax": [
        "error",
        {
          selector:
            "Literal[value=/\\b(?:bg|text|border|from|to|via)-\\[#(?:[0-9A-Fa-f]{3,8})\\]\\b/]",
          message:
            "Do not use Tailwind arbitrary color values. Use token-backed utilities (e.g. bg-surface-raised).",
        },
        {
          selector:
            "TemplateElement[value.raw=/\\b(?:bg|text|border|from|to|via)-\\[#(?:[0-9A-Fa-f]{3,8})\\]\\b/]",
          message:
            "Do not use Tailwind arbitrary color values. Use token-backed utilities (e.g. text-text-primary).",
        },
      ],
    },
  },
  eslintConfigPrettier,
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
  ]),
]);

export default eslintConfig;
