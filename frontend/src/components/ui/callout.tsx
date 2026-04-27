import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Callout container for contextual information and guidance blocks.
 * Use `Banner` for status updates with live-region semantics.
 */
export function Callout({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      className={cn(
        "rounded-md border border-border-default bg-surface-raised p-4 text-body-md text-text-primary",
        className
      )}
      {...props}
    />
  );
}
