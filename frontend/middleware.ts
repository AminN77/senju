import { NextResponse, type NextRequest } from "next/server";

const REFRESH_COOKIE_NAME = process.env.AUTH_REFRESH_COOKIE_NAME ?? "refresh_token";

const PROTECTED_PATH_PREFIXES = ["/dashboard", "/jobs", "/variants", "/upload"];

export function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;

  if (!isProtectedPath(pathname)) {
    return NextResponse.next();
  }

  if (request.cookies.get(REFRESH_COOKIE_NAME)?.value) {
    return NextResponse.next();
  }

  const loginUrl = new URL("/login", request.url);
  const suffix = request.nextUrl.search;
  const nextPath = suffix ? `${pathname}${suffix}` : pathname;
  loginUrl.searchParams.set("next", nextPath);
  return NextResponse.redirect(loginUrl);
}

function isProtectedPath(pathname: string): boolean {
  return PROTECTED_PATH_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)
  );
}

export const config = {
  matcher: ["/dashboard/:path*", "/jobs/:path*", "/variants/:path*", "/upload/:path*"],
};
