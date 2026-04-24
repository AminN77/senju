import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axe } from "../../../tests/utils/axe";
import { render } from "../../../tests/utils/render";
import { AuthenticatedShell } from "./authenticated-shell";

const pushMock = vi.fn();
let pathname = "/dashboard";

vi.mock("next/navigation", () => ({
  usePathname: () => pathname,
  useRouter: () => ({
    push: pushMock,
  }),
}));

vi.mock("@/lib/auth/session-provider", () => ({
  useSession: () => ({
    signOut: async () => {},
  }),
}));

describe("AuthenticatedShell", () => {
  beforeEach(() => {
    pathname = "/dashboard";
    pushMock.mockReset();
  });

  it("has no a11y violations", async () => {
    const { container } = render(
      <AuthenticatedShell>
        <section>
          <h1>Dashboard</h1>
          <p>Overview content</p>
        </section>
      </AuthenticatedShell>
    );

    expect(await axe(container)).toHaveNoViolations();
  });

  it("opens and closes shortcut dialog with keyboard", async () => {
    const user = userEvent.setup();

    render(
      <AuthenticatedShell>
        <section>
          <h1>Dashboard</h1>
        </section>
      </AuthenticatedShell>
    );

    await user.keyboard("?");
    expect(screen.getByRole("dialog", { name: "Keyboard shortcuts" })).toBeInTheDocument();

    await user.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.queryByRole("dialog", { name: "Keyboard shortcuts" })).not.toBeInTheDocument();
    });
  });

  it("navigates by keyboard chord and focuses heading after route change", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <AuthenticatedShell>
        <section>
          <h1>Dashboard</h1>
        </section>
      </AuthenticatedShell>
    );

    await user.keyboard("g");
    await user.keyboard("j");
    expect(pushMock).toHaveBeenCalledWith("/jobs");

    pathname = "/jobs";
    rerender(
      <AuthenticatedShell>
        <section>
          <h1>Jobs</h1>
        </section>
      </AuthenticatedShell>
    );

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Jobs" })).toHaveFocus();
    });
  });
});
