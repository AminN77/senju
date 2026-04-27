import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Inline status banner for page-level notices.
 * Use `Toaster` for transient notifications.
 * A11y: defaults to polite status announcements unless overridden.
 */
export function Banner({
  className,
  role = "status",
  ...props
}: React.ComponentProps<"div"> & { role?: "status" | "alert" }) {
  return (
    <div
      role={role}
      className={cn(
        "rounded-md border border-border-default bg-surface-brand-subtle px-4 py-3 text-body-md text-text-primary",
        className
      )}
      {...props}
    />
  );
}
