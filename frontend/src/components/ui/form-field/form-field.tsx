import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Layout wrapper for label/control/hint/error composition in forms.
 * Use this as the shared shell around all input-like primitives.
 * A11y: consumer wires `htmlFor`, `aria-describedby`, and invalid state on controls.
 */
export function FormField({ className, ...props }: React.ComponentProps<"div">) {
  return <div className={cn("space-y-1.5", className)} {...props} />;
}

/**
 * Supplementary helper text for a form control.
 */
export function FormHint({ className, ...props }: React.ComponentProps<"p">) {
  return <p className={cn("text-body-sm text-text-muted", className)} {...props} />;
}

/**
 * Validation message for invalid form controls.
 */
export function FormError({ className, ...props }: React.ComponentProps<"p">) {
  return <p className={cn("text-body-sm text-danger-solid", className)} role="alert" {...props} />;
}
