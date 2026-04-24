"use client";

import { useState } from "react";
import Link from "next/link";

import { Menu } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ThemeToggle } from "./theme-toggle";

interface AppTopBarProps {
  breadcrumb: string[];
  onOpenMobileSidebar: () => void;
  onOpenShortcuts: () => void;
  onSignOut: () => Promise<void>;
}

export function AppTopBar({
  breadcrumb,
  onOpenMobileSidebar,
  onOpenShortcuts,
  onSignOut,
}: AppTopBarProps) {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <header className="sticky top-0 z-20 border-b border-border-default bg-surface-raised/95 backdrop-blur">
      <div className="flex h-14 items-center justify-between gap-3 px-4">
        <div className="flex items-center gap-3">
          <Button
            type="button"
            size="icon"
            variant="ghost"
            className="lg:hidden"
            onClick={onOpenMobileSidebar}
            aria-label="Open sidebar"
          >
            <Menu className="h-4 w-4" />
          </Button>
          <nav aria-label="Breadcrumb">
            <ol className="flex items-center gap-2 text-body-sm text-text-secondary">
              {breadcrumb.map((item, index) => (
                <li key={`${item}-${index}`}>
                  <span>{item}</span>
                  {index < breadcrumb.length - 1 ? <span className="ml-2">/</span> : null}
                </li>
              ))}
            </ol>
          </nav>
        </div>

        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onOpenShortcuts}
            aria-label="Open global search"
          >
            Search <kbd className="font-mono text-body-sm text-text-secondary">/</kbd>
          </Button>

          <div className="relative">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              aria-label="Open user menu"
              aria-expanded={menuOpen}
              onClick={() => setMenuOpen((value) => !value)}
            >
              User
            </Button>
            {menuOpen ? (
              <div className="absolute right-0 mt-2 w-44 rounded-md border border-border-default bg-surface-raised p-2 shadow-1">
                <Link
                  href="/settings/profile"
                  className="block rounded px-2 py-1 text-body-sm text-text-primary hover:bg-surface-sunken"
                  onClick={() => setMenuOpen(false)}
                >
                  Profile
                </Link>
                <div className="my-1">
                  <ThemeToggle compact />
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start"
                  onClick={() => {
                    setMenuOpen(false);
                    void onSignOut();
                  }}
                >
                  Sign out
                </Button>
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </header>
  );
}
