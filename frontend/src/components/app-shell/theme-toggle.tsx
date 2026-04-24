"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { Button } from "@/components/ui/button";

type ThemeName = "light" | "dark";

const STORAGE_KEY = "senju:theme-preference";

interface ThemeToggleProps {
  compact?: boolean;
}

export function ThemeToggle({ compact = false }: ThemeToggleProps) {
  const [theme, setTheme] = useState<ThemeName>(() => {
    const storage = getStorage();
    if (!storage) {
      return "dark";
    }
    const stored = storage.getItem(STORAGE_KEY);
    return stored === "light" || stored === "dark" ? stored : "dark";
  });

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
  }, [theme]);

  const persistPreference = useCallback(async (nextTheme: ThemeName) => {
    try {
      await fetch("/settings/preferences", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ theme: nextTheme }),
      });
    } catch {
      const storage = getStorage();
      storage?.setItem(STORAGE_KEY, nextTheme);
    }
  }, []);

  const onToggle = useCallback(async () => {
    const nextTheme: ThemeName = theme === "dark" ? "light" : "dark";
    setTheme(nextTheme);
    document.documentElement.setAttribute("data-theme", nextTheme);
    const storage = getStorage();
    storage?.setItem(STORAGE_KEY, nextTheme);
    await persistPreference(nextTheme);
  }, [persistPreference, theme]);

  const label = useMemo(() => {
    if (compact) {
      return theme === "dark" ? "Dark" : "Light";
    }

    return theme === "dark" ? "Switch to light" : "Switch to dark";
  }, [compact, theme]);

  return (
    <Button
      type="button"
      variant="ghost"
      size={compact ? "sm" : "default"}
      onClick={() => void onToggle()}
      aria-label="Toggle theme"
    >
      {label}
    </Button>
  );
}

function getStorage(): Storage | null {
  if (typeof window === "undefined") {
    return null;
  }

  const candidate = window.localStorage;
  if (!candidate || typeof candidate.getItem !== "function" || typeof candidate.setItem !== "function") {
    return null;
  }

  return candidate;
}
