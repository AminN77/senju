import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { SessionProvider, useSession } from "./session-provider";

function SessionProbe() {
  const session = useSession();

  return (
    <section>
      <p data-testid="status">{session.status}</p>
      <p data-testid="user">{session.user?.email ?? "none"}</p>
      <button onClick={() => void session.refresh()} type="button">
        refresh
      </button>
    </section>
  );
}

describe("<SessionProvider>", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("refreshes session and authenticates user", async () => {
    global.fetch = vi.fn(async () => {
      return new Response(
        JSON.stringify({
          access_token: "access-token",
          user: {
            id: "user-1",
            email: "admin@senju.dev",
            role: "admin",
          },
        }),
        {
          status: 200,
          headers: { "content-type": "application/json" },
        }
      );
    }) as typeof fetch;

    render(
      <SessionProvider bootstrap={false}>
        <SessionProbe />
      </SessionProvider>
    );

    expect(screen.getByTestId("status")).toHaveTextContent("unauthenticated");
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));

    await waitFor(() => {
      expect(screen.getByTestId("status")).toHaveTextContent("authenticated");
      expect(screen.getByTestId("user")).toHaveTextContent("admin@senju.dev");
    });
  });
});
