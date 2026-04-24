import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { RequireSession } from "./require-session";
import { SessionProvider } from "./session-provider";

const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  usePathname: () => "/dashboard",
  useSearchParams: () =>
    new URLSearchParams({
      view: "summary",
    }),
}));

describe("<RequireSession>", () => {
  beforeEach(() => {
    replace.mockReset();
  });

  it("renders fallback and redirects unauthenticated users", async () => {
    render(
      <SessionProvider bootstrap={false}>
        <RequireSession fallback={<p>Checking session...</p>}>
          <p>Private content</p>
        </RequireSession>
      </SessionProvider>
    );

    expect(screen.getByText("Checking session...")).toBeInTheDocument();

    await waitFor(() => {
      expect(replace).toHaveBeenCalledWith("/login?next=%2Fdashboard%3Fview%3Dsummary");
    });
  });
});
