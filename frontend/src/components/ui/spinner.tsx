import { Loader2 } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Loading spinner primitive honoring reduced-motion behavior.
 * A11y: pair with text for non-visual loading context.
 */
export function Spinner({ className, ...props }: React.ComponentProps<typeof Loader2>) {
  return (
    <Loader2
      className={cn("h-4 w-4 animate-spin motion-reduce:animate-none", className)}
      {...props}
    />
  );
}
