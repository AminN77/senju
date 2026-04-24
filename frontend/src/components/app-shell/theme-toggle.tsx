"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { createApiClient } from "@/lib/api/client";
import { Button } from "@/components/ui/button";

type ThemeName = "light" | "dark";

const STORAGE_KEY = "senju:theme-preference";
const apiClient = createApiClient();

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
      await apiClient.POST("/v1/settings/preferences", {
        body: { theme: nextTheme },
        headers: { "content-type": "application/json" },
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

  try {
    const candidate = window.localStorage;
    if (
      !candidate ||
      typeof candidate.getItem !== "function" ||
      typeof candidate.setItem !== "function"
    ) {
      return null;
    }

    return candidate;
  } catch {
    return null;
  }
}
