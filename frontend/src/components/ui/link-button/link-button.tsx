import Link from "next/link";
import * as React from "react";

import { Button, type ButtonProps } from "@/components/ui/button";

/**
 * Link-styled button primitive for navigational actions.
 * Use `Button` for in-place actions and mutations.
 * A11y: renders semantic anchor navigation via Next.js `Link`.
 */
export function LinkButton({
  href,
  children,
  ...props
}: Omit<ButtonProps, "asChild"> & { href: string; children: React.ReactNode }) {
  return (
    <Button asChild {...props}>
      <Link href={href}>{children}</Link>
    </Button>
  );
}
