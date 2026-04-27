"use client";

import Link from "next/link";
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

/**
 * Link-styled button primitive for navigational actions.
 * Use `Button` for in-place actions and mutations.
 * A11y: renders semantic anchor navigation via Next.js `Link`.
 */
export function LinkButton({
  href,
  children,
  variant = "default",
  size = "default",
  disabled = false,
  className,
  ...props
}: Omit<React.ComponentProps<typeof Link>, "href" | "className"> &
  VariantProps<typeof linkButtonVariants> & {
    href: string;
    children: React.ReactNode;
    disabled?: boolean;
    className?: string;
  }) {
  const classes = cn(linkButtonVariants({ variant, size }), className);

  if (disabled) {
    return (
      <span aria-disabled className={cn(classes, "pointer-events-none opacity-50")}>
        {children}
      </span>
    );
  }

  return (
    <Link href={href} className={classes} {...props}>
      {children}
    </Link>
  );
}

const linkButtonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-md text-body-md font-medium transition-colors motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-border-focus focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-brand-500 text-text-on-accent hover:opacity-90",
        secondary:
          "bg-surface-brand-subtle text-text-primary border border-border-default hover:bg-surface-raised",
        outline:
          "border border-border-default bg-surface-raised text-text-primary hover:bg-surface-sunken",
        ghost: "text-text-primary hover:bg-surface-sunken",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 px-3",
        lg: "h-11 px-8",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);
