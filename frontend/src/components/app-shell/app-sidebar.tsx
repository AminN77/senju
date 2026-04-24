"use client";

import Link from "next/link";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const navItems = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/projects", label: "Projects" },
  { href: "/jobs", label: "Jobs" },
  { href: "/variants", label: "Variants" },
  { href: "/ml/impact", label: "ML Impact" },
  { href: "/admin", label: "Admin" },
  { href: "/settings", label: "Settings" },
] as const;

interface AppSidebarProps {
  collapsed: boolean;
  mobileOpen: boolean;
  onToggleCollapsed: () => void;
  onCloseMobile: () => void;
}

export function AppSidebar({
  collapsed,
  mobileOpen,
  onToggleCollapsed,
  onCloseMobile,
}: AppSidebarProps) {
  return (
    <>
      {mobileOpen ? (
        <button
          type="button"
          aria-label="Close sidebar overlay"
          className="fixed inset-0 z-30 bg-surface-sunken/60 lg:hidden"
          onClick={onCloseMobile}
        />
      ) : null}

      <aside
        className={cn(
          "fixed left-0 top-0 z-40 flex h-full flex-col border-r border-border-default bg-surface-raised p-3 transition-transform lg:static lg:translate-x-0",
          collapsed ? "w-20" : "w-72",
          mobileOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        )}
        aria-label="Primary navigation"
      >
        <div className="mb-4 flex items-center justify-between gap-2">
          <span className={cn("text-body-sm text-text-secondary", collapsed && "sr-only")}>
            Workspace
          </span>
          <select
            id="workspace-select"
            aria-label="Workspace selector"
            className={cn(
              "h-9 rounded-md border border-border-default bg-surface-sunken px-2 text-body-sm text-text-primary",
              collapsed && "hidden"
            )}
            defaultValue="default"
          >
            <option value="default">Default workspace</option>
          </select>
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={onToggleCollapsed}
            aria-label="Collapse sidebar"
          >
            {collapsed ? ">" : "<"}
          </Button>
        </div>

        <nav className="flex flex-1 flex-col gap-1">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "rounded-md px-3 py-2 text-body-md text-text-primary transition-colors hover:bg-surface-sunken focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus",
                collapsed && "px-2 text-body-sm"
              )}
              onClick={onCloseMobile}
            >
              {collapsed ? item.label.slice(0, 1) : item.label}
            </Link>
          ))}
        </nav>
      </aside>
    </>
  );
}
