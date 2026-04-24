"use client";

import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";

import { useSession } from "@/lib/auth/session-provider";

import { AppSidebar } from "./app-sidebar";
import { AppTopBar } from "./app-topbar";
import { ShortcutDialog } from "./shortcut-dialog";

const shortcutRouteMap: Record<string, string> = {
  d: "/dashboard",
  j: "/jobs",
  p: "/projects",
  v: "/variants",
};

export function AuthenticatedShell({ children }: { children: ReactNode }) {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false);
  const [shortcutOpen, setShortcutOpen] = useState(false);
  const liveRegionRef = useRef<HTMLDivElement>(null);
  const pendingChordListenerRef = useRef<((event: KeyboardEvent) => void) | null>(null);
  const pathname = usePathname() ?? "/dashboard";
  const router = useRouter();
  const session = useSession();

  const breadcrumb = useMemo(
    () =>
      pathname
        .split("/")
        .filter(Boolean)
        .map((segment) =>
          segment.replace(/-/g, " ").replace(/\b\w/g, (char) => char.toUpperCase())
        ),
    [pathname]
  );

  useEffect(() => {
    const heading = document.querySelector("main h1");
    let announcement = document.title || "page";

    if (heading instanceof HTMLElement) {
      heading.tabIndex = -1;
      heading.focus();
      announcement = heading.textContent?.trim() || announcement;
    }

    if (liveRegionRef.current) {
      liveRegionRef.current.textContent = `Navigated to ${announcement}`;
    }
  }, [pathname]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      const isTyping =
        target instanceof HTMLInputElement ||
        target instanceof HTMLTextAreaElement ||
        target?.isContentEditable === true;
      if (isTyping) {
        return;
      }

      if (event.key === "/") {
        event.preventDefault();
        setShortcutOpen(true);
        return;
      }

      if (event.key === "?") {
        event.preventDefault();
        setShortcutOpen(true);
        return;
      }

      if (event.key === "Escape") {
        setShortcutOpen(false);
        setMobileSidebarOpen(false);
        return;
      }

      if (event.key.toLowerCase() === "g") {
        if (pendingChordListenerRef.current) {
          window.removeEventListener("keydown", pendingChordListenerRef.current);
        }

        const onNext = (nextEvent: KeyboardEvent) => {
          const destination = shortcutRouteMap[nextEvent.key.toLowerCase()];
          if (destination) {
            router.push(destination);
          }
          window.removeEventListener("keydown", onNext);
          pendingChordListenerRef.current = null;
        };
        pendingChordListenerRef.current = onNext;
        window.addEventListener("keydown", onNext, { once: true });
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      if (pendingChordListenerRef.current) {
        window.removeEventListener("keydown", pendingChordListenerRef.current);
        pendingChordListenerRef.current = null;
      }
    };
  }, [router]);

  return (
    <div className="flex min-h-screen bg-surface-sunken">
      <a
        href="#main-content"
        className="sr-only z-50 rounded-md bg-surface-raised px-3 py-2 text-text-primary focus:not-sr-only focus:fixed focus:left-3 focus:top-3"
      >
        Skip to main content
      </a>

      <AppSidebar
        collapsed={sidebarCollapsed}
        mobileOpen={mobileSidebarOpen}
        onToggleCollapsed={() => setSidebarCollapsed((value) => !value)}
        onCloseMobile={() => setMobileSidebarOpen(false)}
      />

      <div className="flex min-w-0 flex-1 flex-col">
        <AppTopBar
          breadcrumb={breadcrumb.length > 0 ? breadcrumb : ["Dashboard"]}
          onOpenMobileSidebar={() => setMobileSidebarOpen(true)}
          onOpenShortcuts={() => setShortcutOpen(true)}
          onSignOut={session.signOut}
        />
        <main id="main-content" className="flex-1 p-6">
          {children}
        </main>
      </div>

      <ShortcutDialog open={shortcutOpen} onClose={() => setShortcutOpen(false)} />
      <div aria-live="polite" className="sr-only" ref={liveRegionRef} />
    </div>
  );
}
