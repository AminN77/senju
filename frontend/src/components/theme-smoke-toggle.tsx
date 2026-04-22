"use client";

import { useCallback, useEffect } from "react";

import { Button } from "@/components/ui/button";

const THEMES = ["dark", "light"] as const;
type ThemeName = (typeof THEMES)[number];

const DEFAULT_ROOT_THEME: ThemeName = "dark";

/**
 * Sets `data-theme` on `<html>` for smoke previews. Theming must not be applied
 * on arbitrary elements — only the document element per design system rules.
 */
export function ThemeSmokeToggle() {
  const setTheme = useCallback((t: ThemeName) => {
    document.documentElement.setAttribute("data-theme", t);
  }, []);

  useEffect(
    () => () => {
      document.documentElement.setAttribute("data-theme", DEFAULT_ROOT_THEME);
    },
    []
  );

  return (
    <div
      className="flex flex-wrap items-center gap-2"
      role="group"
      aria-label="Preview theme (sets html data-theme)"
    >
      {THEMES.map((t) => (
        <Button key={t} type="button" size="sm" variant="secondary" onClick={() => setTheme(t)}>
          Preview {t}
        </Button>
      ))}
    </div>
  );
}
