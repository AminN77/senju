import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Tiny status indicator dot intended to pair with text labels/icons.
 * A11y: set `aria-hidden` when decorative.
 */
export function StatusDot({
  className,
  tone = "info",
  ...props
}: React.ComponentProps<"span"> & { tone?: "success" | "warning" | "danger" | "info" | "muted" }) {
  const toneClasses: Record<NonNullable<typeof tone>, string> = {
    success: "bg-success-solid",
    warning: "bg-warning-solid",
    danger: "bg-danger-solid",
    info: "bg-info-solid",
    muted: "bg-neutral-500",
  };
  return (
    <span
      className={cn("inline-block h-2 w-2 rounded-full", toneClasses[tone], className)}
      {...props}
    />
  );
}
