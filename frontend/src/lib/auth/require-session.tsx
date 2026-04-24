"use client";

import { useEffect, type ReactNode } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { useSession } from "./session-provider";
import { buildLoginRedirect } from "./session";

export function RequireSession({
  children,
  fallback = null,
}: {
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const session = useSession();
  const router = useRouter();
  const pathname = usePathname() ?? "/";
  const searchParams = useSearchParams();

  useEffect(() => {
    if (session.status !== "unauthenticated") {
      return;
    }

    const suffix = searchParams.toString();
    const currentPath = suffix ? `${pathname}?${suffix}` : pathname;
    router.replace(buildLoginRedirect(currentPath));
  }, [pathname, router, searchParams, session.status]);

  if (session.status === "authenticated") {
    return <>{children}</>;
  }

  return <>{fallback}</>;
}
