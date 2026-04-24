import { cookies } from "next/headers";
import { redirect } from "next/navigation";

import type { SessionUser } from "./types";

const AUTH_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

interface SessionResponse {
  user: SessionUser;
}

interface RequireSessionOptions {
  getCookieHeader?: () => Promise<string>;
  fetchImpl?: typeof fetch;
  redirectImpl?: (location: string) => never;
}

export async function requireSession(
  currentPath: string,
  options: RequireSessionOptions = {}
): Promise<SessionUser> {
  const getCookieHeader = options.getCookieHeader ?? defaultCookieHeader;
  const fetchImpl = options.fetchImpl ?? fetch;
  const redirectImpl = options.redirectImpl ?? redirect;

  const cookieHeader = await getCookieHeader();
  const response = await fetchImpl(`${AUTH_BASE_URL}/v1/auth/session`, {
    method: "GET",
    cache: "no-store",
    headers: cookieHeader ? { cookie: cookieHeader } : undefined,
  });

  if (response.status === 401) {
    redirectImpl(buildLoginRedirect(currentPath));
  }

  if (!response.ok) {
    throw new Error(`Unable to resolve session (${response.status})`);
  }

  const payload = (await response.json()) as SessionResponse;
  return payload.user;
}

export function buildLoginRedirect(currentPath: string): string {
  const safePath = currentPath.startsWith("/") ? currentPath : `/${currentPath}`;
  const next = encodeURIComponent(safePath);
  return `/login?next=${next}`;
}

async function defaultCookieHeader(): Promise<string> {
  const store = await cookies();
  return store.toString();
}
