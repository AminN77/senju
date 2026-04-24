import { describe, expect, it, vi } from "vitest";

import { axe } from "../../../tests/utils/axe";
import { render } from "../../../tests/utils/render";
import { AuthenticatedShell } from "./authenticated-shell";

vi.mock("next/navigation", () => ({
  usePathname: () => "/dashboard",
  useRouter: () => ({
    push: () => {},
  }),
}));

vi.mock("@/lib/auth/session-provider", () => ({
  useSession: () => ({
    signOut: async () => {},
  }),
}));

describe("AuthenticatedShell", () => {
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
});
