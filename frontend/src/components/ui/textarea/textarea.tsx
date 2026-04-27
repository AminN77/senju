import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Multi-line text input primitive for form content that exceeds one line.
 * Use `Input` for short single-line values.
 * A11y: consumer must provide a visible label or an `aria-label`.
 */
export const Textarea = React.forwardRef<HTMLTextAreaElement, React.ComponentProps<"textarea">>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        ref={ref}
        className={cn(
          "flex min-h-24 w-full rounded-md border border-border-default bg-surface-raised px-3 py-2 text-body-md text-text-primary placeholder:text-text-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        {...props}
      />
    );
  }
);

Textarea.displayName = "Textarea";
