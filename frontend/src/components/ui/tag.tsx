import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Compact metadata tag primitive for categorical labels.
 */
export function Tag({ className, ...props }: React.ComponentProps<"span">) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-sm border border-border-default bg-surface-sunken px-2 py-1 text-caption text-text-secondary",
        className
      )}
      {...props}
    />
  );
}
