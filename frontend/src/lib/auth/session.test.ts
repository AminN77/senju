import { describe, expect, it, vi } from "vitest";

import { buildLoginRedirect, requireSession } from "./session";

describe("requireSession", () => {
  it("redirects to login when unauthenticated", async () => {
    const redirectImpl = vi.fn((location: string) => {
      throw new Error(`redirect:${location}`);
    });

    await expect(
      requireSession("/dashboard", {
        getCookieHeader: async () => "",
        fetchImpl: vi.fn(
          async () => new Response(null, { status: 401 })
        ) as unknown as typeof fetch,
        redirectImpl,
      })
    ).rejects.toThrowError("redirect:/login?next=%2Fdashboard");

    expect(redirectImpl).toHaveBeenCalledWith("/login?next=%2Fdashboard");
  });

  it("returns authenticated user", async () => {
    const user = {
      id: "user-1",
      email: "analyst@senju.dev",
      role: "analyst" as const,
    };

    const resolved = await requireSession("/dashboard", {
      getCookieHeader: async () => "refresh_token=abc",
      fetchImpl: vi.fn(
        async () =>
          new Response(JSON.stringify({ user }), {
            status: 200,
            headers: { "content-type": "application/json" },
          })
      ) as unknown as typeof fetch,
      redirectImpl: ((location: string) => {
        throw new Error(`unexpected redirect:${location}`);
      }) as never,
    });

    expect(resolved).toEqual(user);
  });
});

describe("buildLoginRedirect", () => {
  it("normalizes and encodes next path", () => {
    expect(buildLoginRedirect("dashboard?tab=jobs")).toBe("/login?next=%2Fdashboard%3Ftab%3Djobs");
  });
});
