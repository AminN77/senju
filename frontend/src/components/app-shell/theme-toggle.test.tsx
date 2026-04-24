import { fireEvent, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { render } from "../../../tests/utils/render";
import { ThemeToggle } from "./theme-toggle";

describe("ThemeToggle", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("persists and applies theme preference", () => {
    render(<ThemeToggle />);
    const toggle = screen.getByRole("button", { name: "Toggle theme" });
    fireEvent.click(toggle);

    expect(document.documentElement).toHaveAttribute("data-theme", "light");
  });
});
