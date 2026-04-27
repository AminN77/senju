import * as React from "react";

import { Button, type ButtonProps } from "@/components/ui/button";

/**
 * Icon-only action button with enforced icon sizing and button semantics.
 * Use `Button` when visible text is required.
 * A11y: consumer must provide `aria-label` describing the action.
 */
export function IconButton({
  className,
  size = "icon",
  children,
  ...props
}: Omit<ButtonProps, "children"> & { children: React.ReactNode }) {
  return (
    <Button size={size} className={className} {...props}>
      {children}
    </Button>
  );
}
