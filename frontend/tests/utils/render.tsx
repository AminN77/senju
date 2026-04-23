import { render as rtlRender, type RenderOptions, type RenderResult } from "@testing-library/react";
import type { ReactElement } from "react";

type ThemeName = "light" | "dark";

interface RenderWithProvidersOptions extends Omit<RenderOptions, "wrapper"> {
  theme?: ThemeName;
}

/**
 * Shared test render helper to enforce theme/token context consistently.
 */
export function render(
  ui: ReactElement,
  { theme = "dark", ...options }: RenderWithProvidersOptions = {}
): RenderResult {
  document.documentElement.setAttribute("data-theme", theme);

  return rtlRender(<div data-theme={theme}>{ui}</div>, options);
}
