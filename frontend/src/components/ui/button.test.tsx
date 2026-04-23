import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Button } from "@/components/ui/button";
import { axe } from "../../../tests/utils/axe";
import { render } from "../../../tests/utils/render";

describe("Button", () => {
  it("renders in both light and dark themes", () => {
    const { rerender } = render(<Button>Run smoke action</Button>, { theme: "light" });
    expect(screen.getByRole("button", { name: "Run smoke action" })).toBeInTheDocument();
    expect(document.documentElement).toHaveAttribute("data-theme", "light");

    rerender(<Button>Run smoke action</Button>);
    document.documentElement.setAttribute("data-theme", "dark");
    expect(document.documentElement).toHaveAttribute("data-theme", "dark");
  });

  it("has no a11y violations in both themes", async () => {
    const { container, rerender } = render(<Button>Run smoke action</Button>, { theme: "light" });
    expect(await axe(container)).toHaveNoViolations();

    document.documentElement.setAttribute("data-theme", "dark");
    rerender(<Button>Run smoke action</Button>);
    expect(await axe(container)).toHaveNoViolations();
  });
});
